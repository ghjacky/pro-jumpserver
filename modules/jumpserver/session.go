package jumpserver

import (
	"fmt"
	"github.com/gliderlabs/ssh"
	ssh2 "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"zeus/common"
	"zeus/models"
	"zeus/modules/assets"
	"zeus/modules/audit"
)

const (
	SessionRedisPrefix                = "zeus_jump_session"
	SessionNormalEventsStoreKeyPrefix = "session_NM"
	SessionKBEventsStoreKeyPrefix     = "session_KB"
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
	assets         []*models.Server
	searchedAssets []*models.Server
	Banner
	selectedIDC   string
	sessionID     string
	userIP        string
	jpsIP         string
	serverIP      string
	kbEventWriter *audit.ChanWriter
}

// sessionHandler handle user connection when connecting to jumpserver
func sessionHandler(session ssh.Session) {
	defer func() {
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
	term := NewTerminal(wrapperSess, SessionTerminalPrompt)
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
	h.kbEventWriter = audit.NewChanWriter()
	banner := newDefaultBanner()
	h.Banner = banner
	h.displayBanner()
	h.winWatchChan = make(chan bool)
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
			if err == nil && idcID >= 0 && idcID < len(IDCS) {
				h.selectedIDC = IDCS[idcID]
				h.Banner.setMainMenu(IDCS[idcID])
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
		case 0, 1:
			switch strings.ToLower(line) {
			case "", "p":
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
	h.assets = assets.FetchPermedAssets(user, h.selectedIDC)
}

func (h *interactiveHandler) displayAllAssets() {
	for _, a := range h.assets {
		_, _ = h.term.c.Write([]byte(fmt.Sprintf("%d	%s		%s	%s\n", a.ID, a.Hostname, a.IP, a.IDC)))
	}
}

func (h *interactiveHandler) searchAssets(pattern string) {
	h.searchedAssets = []*models.Server{}
	for _, a := range h.assets {
		if strings.Contains(a.Hostname, pattern) || strings.Contains(a.IP, pattern) || strings.Contains(fmt.Sprintf("%d", a.ID), pattern) {
			h.searchedAssets = append(h.searchedAssets, a)
		}
	}
	// 如果只匹配到一个主机，则直接登陆，两个及以上主机则返回列表展示
	if len(h.searchedAssets) == 1 {
		a := h.searchedAssets[0]
		h.serverIP = a.IP
		switch a.Type {
		case models.AssetTypeSSH:
			as := &assets.ASSH{}
			as.IP = a.IP
			as.PORT = a.Port
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
		h.displaySearchedAssets()
	}
}

func (h *interactiveHandler) displaySearchedAssets() {
	for _, a := range h.searchedAssets {
		_, _ = h.term.c.Write([]byte(fmt.Sprintf("%d	%s		%s	%s\n", a.ID, a.Hostname, a.IP, a.IDC)))
	}
}

func (h *interactiveHandler) Terminal(session *ssh2.Session) (err error) {
	var watcherExitChan = make(chan bool, 0)
	var sessionExitChan = make(chan bool, 0)
	modes := ssh2.TerminalModes{
		ssh2.ECHO:          1,
		ssh2.ECHOCTL:       0,
		ssh2.TTY_OP_ISPEED: 14400,
		ssh2.TTY_OP_OSPEED: 14400,
	}
	// 生成资产登陆事件并存储
	h.generateServerLoginEvent()
	// 开始监控keyboard single character事件, 并存储，用户后续播放
	h.WatchKBEvent(session, sessionExitChan, watcherExitChan)
	//session.Stdout = h.term.c
	session.Stdin = h.term.c
	session.Stderr = h.term.c
	termFD := int(os.Stdin.Fd())
	width, height := h.term.GetSize()
	termState, _ := terminal.MakeRaw(termFD)
	defer func() {
		if err := terminal.Restore(termFD, termState); err != nil {
			common.Log.Errorln("Couldn't restore original terminal")
		}
	}()
	err = session.RequestPty("xterm-256color", height, width, modes)
	err = session.Shell()
	err = session.Wait()
	// session退出后通过发送退出信号，关闭相应goroutine和io等
	sessionExitChan <- true
	watcherExitChan <- true
	return
}
