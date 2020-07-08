package jumpserver

import (
	"fmt"
	"github.com/gliderlabs/ssh"
	ssh2 "golang.org/x/crypto/ssh"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
	"zeus/common"
	"zeus/models"
	"zeus/modules/assets"
	"zeus/modules/webserver/users"
)

const (
	SessionRedisPrefix                 = "zeus_jump_session"
	SessionNormalEventsStoreKeyPrefix  = "session_NM"
	SessionKBEventsStoreKeyPrefix      = "session_KB"
	SessionCommandEventsStoreKeyPrefix = "session_CMD"
	//SessionEventBufferFlushInterval   = 100 * time.Millisecond
	SessionTerminalPrompt = "JPS >> "
)

type interactiveHandler struct {
	sess         *WrapperSession
	user         string
	term         *Terminal
	winWatchChan chan bool
	mu           *sync.RWMutex
	//nodeDataLoaded  chan struct{}
	//assetDataLoaded chan struct{}
	servers         models.Servers
	searchedServers models.Servers
	Banner
	selectedIDC     string
	sessionID       string
	userIP          string
	jpsIP           string
	serverIP        string
	kbEventWriter   *ChanWriter
	execEventWriter *LineChanWriter
	env             map[string]string
}

// sessionHandler handle user connection when connecting to jumpserver
var SessionPool = map[string]map[string]ssh.Session{}

// 登陆后首先更新用户活动状态(db)
func changeOnlineStatus(username string, status string) {
	var u models.User
	if err := common.Mysql.Find(&u, "username = ?", username).Error; err != nil {
		common.Log.Warnf("Couldn't find user: %s in db, maybe new user", username)
		u.Username = username
	}
	u.Active = status
	common.Mysql.Save(&u)
}

func addSessionToPool(session ssh.Session) {
	user := session.Context().Value(ssh.ContextKeyUser).(string)
	sid := session.Context().Value(ssh.ContextKeySessionID).(string)
	lock := sync.Mutex{}
	lock.Lock()
	if _, ok := SessionPool[user]; ok {
		SessionPool[user][sid] = session
	} else {
		SessionPool[user] = map[string]ssh.Session{}
		SessionPool[user][sid] = session
	}
	changeOnlineStatus(user, models.UserActiveYes)
	lock.Unlock()
}
func removeSessionFromPool(session ssh.Session) {
	user := session.Context().Value(ssh.ContextKeyUser).(string)
	sid := session.Context().Value(ssh.ContextKeySessionID).(string)
	lock := sync.Mutex{}
	lock.Lock()
	if _, ok := SessionPool[user]; ok {
		delete(SessionPool[user], sid)
		if len(SessionPool[user]) == 0 {
			delete(SessionPool, user)
		}
	}
	changeOnlineStatus(user, models.UserActiveNo)
	lock.Unlock()
}
func sessionHandlerWrapper(session ssh.Session) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func(session ssh.Session) {
		addSessionToPool(session)
		sessionHandler(session)
		wg.Done()
	}(session)
	wg.Wait()
}
func sessionHandler(session ssh.Session) {
	defer func() {
		removeSessionFromPool(session)
		if err := session.Close(); err != nil {
			common.Log.Warnf("Couldn't close session: %s of user: %s", session)
		}
	}()

	// 返回一个自定义终端
	_, _, ok := session.Pty()
	if ok {
		// 首先创建相关session事件存储目录(文件）

		handler := newInteractiveHandler(session, session.User())
		go handler.fetchPermissionAssets()
		// 初始化handler
		handler.Initial(session)
		// session退出前关闭相应缓冲区
		defer func() {
			// 关闭eventWriter
			if err := handler.kbEventWriter.Close(); err != nil {
				common.Log.Errorln("couldn't close watcher channel")
			} else {
				common.Log.Infoln("closing watcher channel")
			}
		}()
		// 生成用户跳板机登陆事件，并存储
		handler.generateJPSLoginEvent()

		// 监控window size变化
		go handler.watchWinSizeChange()

		// 定义菜单切换及退出信号
		var sessionExitSignal = make(chan bool, 1)
		// jumpserver 终端菜单展示及处理 （此处逻辑后续需要优化以方便添加或减除多级菜单）
		for {
			select {
			case exit := <-sessionExitSignal:
				if exit {
					return
				} else {
					handler.selectIDC(sessionExitSignal) // sessionExitSignal为false则说明是返回上级菜单操作，仅当q命令时为退出信号
				}
			default:
				handler.selectIDC(sessionExitSignal) // 首次terminal初始化过后，sessionExitSignal为空
			}
		}
	} else {
		common.Log.Errorf("Couldn't to request pty for user: %s", session.User())
	}

}

func newInteractiveHandler(sess ssh.Session, user string) *interactiveHandler {
	wrapperSess := NewWrapperSession(sess)
	term := NewTerminal(wrapperSess, SessionTerminalPrompt, 80, 24)
	handler := &interactiveHandler{
		sess: wrapperSess,
		user: user,
		term: term,
		mu:   new(sync.RWMutex),
		//nodeDataLoaded:  make(chan struct{}),
		//assetDataLoaded: make(chan struct{}),
	}
	handler.Initial(sess)
	return handler
}

func (h *interactiveHandler) Initial(session ssh.Session) {
	h.sessionID = session.Context().Value(ssh.ContextKeySessionID).(string)
	h.userIP, _, _ = net.SplitHostPort(session.RemoteAddr().String())
	h.jpsIP, _, _ = net.SplitHostPort(session.LocalAddr().String())
	banner := newDefaultBanner()
	h.Banner = banner
	h.displayBanner()
	h.winWatchChan = make(chan bool)
	h.kbEventWriter = NewChanWriter(h)
	h.execEventWriter = NewLineChanWriter(h)
	h.env = map[string]string{}
	h.parEnv(session.Environ())
}

func (h *interactiveHandler) parEnv(env []string) {
	for _, s := range env {
		kv := strings.Split(s, "=")
		if len(kv) > 1 {
			h.env[kv[0]] = kv[1]
		}
	}
}
func (h *interactiveHandler) Set() {
	h.term.SetPrompt("")
	h.term.moveCursorToPos(0)
	h.term.history.entries = make([]string, h.term.history.max)
	h.term.history.head = 0
	h.term.history.size = 0
}
func (h *interactiveHandler) Reset() {
	h.env = map[string]string{}
	h.winWatchChan = make(chan bool)
	h.term.SetPrompt(SessionTerminalPrompt)
	h.term.history.entries = make([]string, h.term.history.max)
	h.term.history.head = 0
	h.term.history.size = 0
}

func (h *interactiveHandler) displayBanner() {
	displayBanner(h.sess, h.user, h.Banner)
}

func (h *interactiveHandler) watchWinSizeChange() {
	sessChan := h.sess.WinCh()
	winChan := sessChan
	for {
		select {
		case <-h.sess.Sess.Context().Done():
			return
		case sig, ok := <-h.winWatchChan:
			if !ok {
				return
			}
			switch sig {
			case false:
				winChan = nil
			case true:
				winChan = sessChan
			}
		case win, ok := <-winChan:
			if !ok {
				return
			}
			common.Log.Debugf("Term window size change: %d*%d", win.Height, win.Width)
			_ = h.term.SetSize(win.Width, win.Height)
		}
	}
}

func (h *interactiveHandler) pauseWatchWinSize() {
	h.winWatchChan <- false
}

func (h *interactiveHandler) resumeWatchWinSize() {
	h.winWatchChan <- true
}

func (h *interactiveHandler) selectIDC(sessionExitSignal chan bool) {
	for {
		line, err := h.term.ReadLine()
		if err != nil {
			if err != io.EOF {
				common.Log.Debug("User disconnected")
			} else {
				// 检测到EOF就退出
				common.Log.Error("Read from user err: ", err)
				sessionExitSignal <- true
				return
			}
			return
		}
		line = strings.TrimSpace(line)
		switch line {
		case "q":
			sessionExitSignal <- true
			return
		default:
			idcID, err := strconv.Atoi(line)
			if err == nil && idcID >= 0 && idcID < len(IDCs) {
				h.selectedIDC = IDCs[idcID]
				h.Banner.setMainMenu(IDCs[idcID])
				h.displayBanner()
				h.Dispatch(sessionExitSignal)
				return
			} else {
				_, _ = h.term.c.Write([]byte("输入的序号有误，请重新输入！\n"))
			}
		}
	}
}

func (h *interactiveHandler) Dispatch(sessionExitSignal chan bool) {
	for {

		line, err := h.term.ReadLine()

		if err != nil {
			if err != io.EOF {
				common.Log.Debug("User disconnected")
			} else {
				// 检测到EOF就退出
				common.Log.Error("Read from user err: ", err)
				sessionExitSignal <- true
				return
			}
			break
		}
		line = strings.TrimSpace(line)
		//<-h.assetDataLoaded
		switch len(line) {
		case 0:
		case 1:
			switch strings.ToLower(line) {
			case "p":
				h.mu.RLock()
				// 展示资源
				h.displayAllAssets()
				h.mu.RUnlock()
			case "h":
				h.displayBanner()
			case "r":
				h.Banner = newDefaultBanner()
				h.displayBanner()
				sessionExitSignal <- false
				return
				//h.refreshAssetsAndNodesData()
			case "q":
				common.Log.Info("exit session")
				sessionExitSignal <- true
				return
			default:
				// 默认搜索输入的字符串（从id、ip、hostname中搜索字符串）
				h.searchAssets(line)
			}
		default:
			switch {
			case line == "exit", line == "quit":
				common.Log.Info("exit session")
				sessionExitSignal <- true
				return
			default:
				// 默认搜索输入的字符串（从id、ip、hostname中搜索字符串）
				h.searchAssets(line)
			}
		}

	}
}

func (h *interactiveHandler) fetchPermissionAssets() {
	user := models.User{Username: h.user}
	h.servers = users.FilterPermissionServersByIDC(&user, h.selectedIDC)
}

func (h *interactiveHandler) displayAllAssets() {
	for i, s := range h.servers {
		_, _ = h.term.c.Write([]byte(fmt.Sprintf("%d	%s		%s	%s\n", i+1, s.Hostname, s.IP, s.IDC)))
	}
}

func (h *interactiveHandler) searchAssets(pattern string) {
	h.searchedServers = []*models.Server{}
	for i, a := range h.servers {
		if strings.HasPrefix(a.Hostname, pattern) || strings.HasPrefix(a.IP, pattern) || strings.HasPrefix(fmt.Sprintf("%d", i+1), pattern) {
			h.searchedServers = append(h.searchedServers, a)
		}
	}
	// 如果只匹配到一个主机，则直接登陆，两个及以上主机则返回列表展示
	if len(h.searchedServers) == 1 {
		s := h.searchedServers[0]
		h.serverIP = s.IP
		switch s.Type {
		case models.ServerTypeSSH:
			as := &assets.ASSH{}
			as.IP = s.IP
			as.PORT = s.Port
			as.USER = h.user
			// 从session context中获取用户的登陆凭证用于远程主机登陆
			as.PASS = h.sess.Sess.Context().Value("loginPass").(string)
			ias, err := assets.NewAssetClient(as)
			if err != nil {
				_, _ = h.term.c.Write([]byte(err.Error()))
				return
			}
			as = ias.(*assets.ASSH)
			// 建立一个到远端主机到ssh session
			subSession := as.NewSession().(*ssh2.Session)
			defer subSession.Close()
			if subSession != nil {
				if err := h.Terminal(subSession); err != nil {
					common.Log.Errorf("Couldn't connect to host: %s:%d using ssh", as.IP, as.PORT)
					_, _ = h.term.c.Write([]byte(fmt.Sprintf("登陆主机: %s: %d失败\n", as.IP, as.PORT)))
				}
			} else {
				_, _ = h.term.c.Write([]byte(fmt.Sprintf("登陆主机: %s: %d失败\n", as.IP, as.PORT)))
			}
		}
	} else {
		h.displaySearchedServers()
	}
}

func (h *interactiveHandler) displaySearchedServers() {
	for i, a := range h.searchedServers {
		_, _ = h.term.c.Write([]byte(fmt.Sprintf("%d	%s		%s	%s\n", i+1, a.Hostname, a.IP, a.IDC)))
	}
}

var sessionDone = make(chan bool, 0)

func (h *interactiveHandler) Terminal(session *ssh2.Session) (err error) {
	modes := ssh2.TerminalModes{
		ssh2.ECHO:          1,
		ssh2.ECHOCTL:       0,
		ssh2.TTY_OP_ISPEED: 14400,
		ssh2.TTY_OP_OSPEED: 14400,
	}

	// 开始监控keyboard single character事件, 并存储，用户后续播放
	loginServerEventId := h.generateServerLoginEvent()
	var kbWatcherExitChan = make(chan bool, 0)
	go h.WatchKBEvent(kbWatcherExitChan, loginServerEventId)

	// 监控命令执行
	var execWatcherExitChan = make(chan bool, 0)
	go h.WatchExecEvent(execWatcherExitChan, loginServerEventId)

	//session.Stdout = h.term.c
	//session.Stdin = h.term.c
	//session.Stderr = h.term.c
	// 监控h.term输入输出及错误 生成相应事件并存储
	sout, oerr := session.StdoutPipe()
	sin, ierr := session.StdinPipe()
	serr, eerr := session.StderrPipe()
	if oerr != nil || ierr != nil || eerr != nil {
		common.Log.Errorf("session 绑定失败，退出")
		execWatcherExitChan <- true
		kbWatcherExitChan <- true
		return
	}
	// 此处绑定并分流stdin\stdout\stderr
	stdc := make(chan interface{}, 3)
	go func() {
		mw := io.MultiWriter(h.term.c, h.kbEventWriter)
		for {
			select {
			case <-stdc:
				return
			default:
				_, ie := io.Copy(mw, sout)
				if ie != nil {
					common.Log.Errorf("io error (sout): %s", ie.Error())
				}
			}
		}
	}()
	go func() {
		mw := io.MultiWriter(h.term.c, h.kbEventWriter)
		for {
			select {
			case <-stdc:
				return
			default:
				_, oe := io.Copy(mw, serr)
				if oe != nil {
					common.Log.Errorf("io error (serr): %s", oe.Error())
				}
			}
		}
	}()
	go func() {
		mw := io.MultiWriter(h.execEventWriter, sin)
		for {
			select {
			case <-stdc:
				return
			default:
				_, ee := io.Copy(mw, h.term.c)
				if ee != nil {
					common.Log.Errorf("io error (sin): %s", ee.Error())
				}
			}
		}
	}()
	// session结束时终止后台监控任务
	go func() {
		for {
			select {
			case done := <-sessionDone:
				if done {
					common.Log.Infoln("session done")
					stdc <- "sin"
					stdc <- "serr"
					stdc <- "sout"
					execWatcherExitChan <- true
					kbWatcherExitChan <- true
					return
				}
			}
		}
	}()

	width, height := h.term.GetSize()
	//termFD := int(os.Stdin.Fd())
	//termState, _ := terminal.MakeRaw(termFD)
	//defer func() {
	//	if err := terminal.Restore(termFD, termState); err != nil {
	//		common.Log.Errorln("Couldn't restore original terminal")
	//	}
	//}()
	err = session.RequestPty("xterm-256color", height, width, modes)
	// 登陆到远程主机,重置相关项（设置prompt长度为0，以使cursorX表现正常）
	h.Set()
	err = session.Shell()
	err = session.Wait()
	// session退出，重置相关项
	h.Reset()
	// session退出后通过发送退出信号，关闭相应goroutine和io等
	sessionDone <- true
	return
}

func ExitSessionBgTask(millisecond time.Duration) {
	var done chan interface{}
	go func() {
		time.Sleep(millisecond * time.Millisecond)
		done <- 1
	}()
	for {
		select {
		case <-done:
			break
		default:
			sessionDone <- true
			break
		}
	}
}

//type ChanBuffer chan []byte
//
//func (cb ChanBuffer) Read(p []byte) (n int, err error) {
//
//}
//
//func (cb ChanBuffer) Write(p []byte) (n int, err error) {
//
//}
//
//func (cb ChanBuffer) Close() error{
//	close(cb)
//	return nil
//}
