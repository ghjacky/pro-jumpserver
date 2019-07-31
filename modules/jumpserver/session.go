package jumpserver

import (
	"github.com/gliderlabs/ssh"
	"io"
	"net"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
	"zeus/common"
	"zeus/modules/audit"
)

const (
	SessionRedisPrefix                = "zeus_jump_session"
	SessionNormalEventsStoreKeyPrefix = "session_NM"
	SessionKBEventsStoreKeyPrefix     = "session_KB"
	SessionEventBufferFlushInterval   = 100 * time.Millisecond
	SessionTerminalPrompt             = "JPS >> "
)

// sessionHandler handle user connection when connecting to jumpserver
func sessionHandler(session ssh.Session) {
	//wg := sync.WaitGroup{}
	defer func() {
		if err := session.Close(); err != nil {
			common.Log.Warnf("Couldn't close session: %s of user: %s", session)
		}
	}()
	// 用户登陆成功, 获取相关信息，生成登陆事件并存储
	sessionID := session.Context().Value(ssh.ContextKeySessionID)
	timestamp := time.Now().Unix()
	clientIP, _, _ := net.SplitHostPort(session.RemoteAddr().String())
	serverIP, _, _ := net.SplitHostPort(session.LocalAddr().String())
	user := session.User()
	// 生成登陆事件
	event := (audit.NewEvent(audit.EventTypeUserLoginToJPS)).(*audit.LoginEvent)
	// 创建一个缓冲区
	buffer := make(chan []byte, 10240)
	defer close(buffer)
	// 创建两个文件事件存储器（单字符事件和其他分开存储，单字符事件用于后续回放）
	fsNormal := audit.NewStore(audit.StoreFile, path.Join(SessionNormalEventRecordDir, strings.Join([]string{SessionNormalEventsStoreKeyPrefix, sessionID.(string)}, "_")))
	defer func() {
		if err := fsNormal.Close(); err != nil {
			common.Log.Errorln("Couldn't close events store")
		}
	}()
	fsKB := audit.NewStore(audit.StoreFile, path.Join(SessionKBEventRecordDir, strings.Join([]string{SessionKBEventsStoreKeyPrefix, sessionID.(string)}, "_")))
	defer func() {
		if err := fsKB.Close(); err != nil {
			common.Log.Errorln("Couldn't close events store")
		}
	}()
	// 更新登陆事件信息
	event.SessionID = sessionID.(string)
	event.User = user
	event.Timestamp = timestamp
	event.ClientIP = clientIP
	event.ServerIP = serverIP
	event.SetStore(&fsNormal)
	if err := event.WriteToBuffer(event, buffer); err != nil {
		common.Log.Errorf("Failed to write event to buffer")
	}

	// goroutine 后台定时从flush buffer到store
	//wg.Add(1)
	go func() {
		//defer wg.Done()
		event.FlushBuffer(buffer, SessionEventBufferFlushInterval)
	}()

	// 返回一个自定义终端
	_, _, ok := session.Pty()
	if ok {
		handler := newInteractiveHandler(session, user)
		var sessionExitSignal = make(chan bool, 0)
		var mainCliLoopExitSignal = make(chan bool, 0)
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case exit := <-sessionExitSignal:
					if exit {
						close(sessionExitSignal)
						mainCliLoopExitSignal <- true
						return
					}
				}
			}
		}()

		go handler.watchWinSizeChange()
		go func() {
			for {
				select {
				case <-mainCliLoopExitSignal:
					return
				default:
					handler.Banner = newDefaultBanner()
					handler.displayBanner()
					handler.selectIDC(sessionExitSignal)
				}
			}
		}()
		wg.Wait()
	} else {
		common.Log.Errorf("Couldn't to request pty for user: %s", user)
	}

	// 等待goroutine退出
	//wg.Wait()
}

func newInteractiveHandler(sess ssh.Session, user string) *interactiveHandler {
	wrapperSess := NewWrapperSession(sess)
	term := NewTerminal(wrapperSess, SessionTerminalPrompt)
	handler := &interactiveHandler{
		sess:            wrapperSess,
		user:            user,
		term:            term,
		mu:              new(sync.RWMutex),
		nodeDataLoaded:  make(chan struct{}),
		assetDataLoaded: make(chan struct{}),
	}
	handler.Initial()
	return handler
}

type interactiveHandler struct {
	sess            *WrapperSession
	user            string
	term            *Terminal
	winWatchChan    chan bool
	mu              *sync.RWMutex
	nodeDataLoaded  chan struct{}
	assetDataLoaded chan struct{}
	Banner
	selectedIDC string
}

func (h *interactiveHandler) Initial() {
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
	line, err := h.term.ReadLine()
	if err != nil {
		if err != io.EOF {
			common.Log.Debug("User disconnected")
		} else {
			common.Log.Error("Read from user err: ", err)
		}
		return
	}
	line = strings.TrimSpace(line)
	idcID, err := strconv.Atoi(line)
	if err == nil && idcID >= 0 && idcID < len(IDCS) {
		h.selectedIDC = IDCS[idcID]
		h.Banner.setMainMenu(IDCS[idcID])
		h.displayBanner()
		h.Dispatch(sessionExitSignal)
	} else {
		_, _ = h.term.c.Write([]byte("输入的序号有误，请重新输入！\n"))
	}
}

func (h *interactiveHandler) Dispatch(sessionExitSignal chan bool) {
	for {
		line, err := h.term.ReadLine()

		if err != nil {
			if err != io.EOF {
				common.Log.Debug("User disconnected")
			} else {
				common.Log.Error("Read from user err: ", err)
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
				//h.displayAssets(h.assets)
				h.mu.RUnlock()
			//case "g":
			//	<-h.nodeDataLoaded
			//	h.displayNodes(h.nodes)
			case "h":
				h.displayBanner()
			case "r":
				sessionExitSignal <- false
				return
				//h.refreshAssetsAndNodesData()
			case "q":
				common.Log.Info("exit session")
				sessionExitSignal <- true
				return
			default:
				//assets := h.searchAsset(line)
				//h.displayAssetsOrProxy(assets)
			}
		default:
			switch {
			case line == "exit", line == "quit":
				common.Log.Info("exit session")
				return
			case strings.Index(line, "/") == 0:
				//searchWord := strings.TrimSpace(line[1:])
				//assets := h.searchAsset(searchWord)
				//h.displayAssets(assets)
			case strings.Index(line, "g") == 0:
				searchWord := strings.TrimSpace(strings.TrimPrefix(line, "g"))
				if num, err := strconv.Atoi(searchWord); err == nil {
					if num >= 0 {
						//<-h.nodeDataLoaded
						//assets := h.searchNodeAssets(num)
						//h.displayAssets(assets)
						continue
					}
				}
				//assets := h.searchAsset(line)
				//h.displayAssetsOrProxy(assets)
			default:
				//assets := h.searchAsset(line)
				//h.displayAssetsOrProxy(assets)
			}
		}

	}
}
