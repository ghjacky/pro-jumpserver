package jumpserver

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
	"io"
	"path"
	"strings"
	"time"
	"zeus/common"
	"zeus/modules/audit"
)

// 生成指定类型事件
func (h *interactiveHandler) newEvent(t string) (e audit.IEvent) {
	ce := audit.Event{User: h.user, SessionID: h.sessionID}
	switch t {
	case audit.EventTypeUserLoginToJPS:
		e = &audit.LoginToJpsEvent{
			Event: ce,
		}
		e.(*audit.LoginToJpsEvent).Type = audit.EventTypeUserLoginToJPS
	case audit.EventTypeUserLoginToServer:
		e = &audit.LoginToServerEvent{
			Event: ce,
		}
		e.(*audit.LoginToServerEvent).Type = audit.EventTypeUserLoginToServer
	case audit.EventTypeKeyBoardPress:
		e = &audit.KBEvent{
			Event: ce,
		}
		e.(*audit.KBEvent).Type = audit.EventTypeKeyBoardPress
	}
	return
}

// 监听键盘按键事件
func (h *interactiveHandler) WatchKBEvent(session *ssh.Session, sessionDone, watcherDone chan bool, loginServerEventId uuid.UUID) {
	var flushDone = make(chan interface{}, 0)
	// 文件事件存储器
	fsKB := audit.NewStore(audit.StoreFile, path.Join(SessionKBEventRecordDir, strings.Join([]string{SessionKBEventsStoreKeyPrefix, audit.EventTypeKeyBoardPress, h.serverIP, h.sessionID, loginServerEventId.String()}, "_")))

	// 监控h.term输入输出及错误 生成相应事件并存储
	sout, _ := session.StdoutPipe()

	// 监控terminal按键，生成并存储事件
	kbEvent := h.newEvent(audit.EventTypeKeyBoardPress).(*audit.KBEvent)
	kbEvent.ClientIP = h.serverIP
	kbEvent.ServerIP = h.serverIP
	kbEvent.SetStore(&fsKB)
	kbEvent.Buffer = make(chan []byte, 10240)
	// goroutine 后台定时从flush buffer到store
	go kbEvent.FlushBuffer(flushDone, audit.SessionEventBufferFlushInterval)
	//
	go func(se chan bool) {
		defer func() {
			flushDone <- 1
		}()
		for {
			select {
			case done := <-se:
				if done {
					common.Log.Infoln("io copy stopping")
					return
				}
			default:
				mw := io.MultiWriter(h.term.c, h.kbEventWriter)
				_, err := io.Copy(mw, sout)
				if err != nil {
					common.Log.Errorln("io error sout")
				}
			}
		}
	}(sessionDone)
	go h.kbEventWriter.Watch(kbEvent, watcherDone)
}

// 生成jump server登陆事件
//var JpsFlushDone = make(chan interface{}, 0)
func (h *interactiveHandler) generateJPSLoginEvent() {
	// 用户登陆成功, 获取相关信息，生成登陆事件并存储，此时，用户已进入被监控状态
	loginEvent := h.newEvent(audit.EventTypeUserLoginToJPS).(*audit.LoginToJpsEvent)
	loginEvent.Timestamp = time.Now().UnixNano()
	loginEvent.ClientIP = h.userIP
	loginEvent.ServerIP = h.jpsIP
	loginEvent.ID = uuid.New()
	loginEvent.Buffer = make(chan []byte, 1)

	// 创建文件事件存储器（单字符事件和其他分开存储，单字符事件用于后续回放）
	//fsNormal := audit.NewStore(audit.StoreFile, path.Join(SessionNormalEventRecordDir, strings.Join([]string{SessionNormalEventsStoreKeyPrefix, audit.EventTypeUserLoginToJPS, h.userIP, h.sessionID}, "_")))
	//
	//loginEvent.SetStore(&fsNormal)
	//go loginEvent.FlushBuffer(JpsFlushDone, audit.SessionEventBufferFlushInterval)
	// 更新登陆事件信息
	if err := loginEvent.WriteToBuffer(loginEvent); err != nil {
		common.Log.Errorf("Failed to write event to buffer")
	}
	return
}

// 生成远程主机登陆事件
func (h *interactiveHandler) generateServerLoginEvent() (loginServerEventID uuid.UUID) {
	//var flushDone = make(chan interface{}, 0)
	//defer func() {
	//	flushDone <- 1
	//}()
	// 用户登陆资产成功, 获取相关信息，生成登陆事件并存储，此时，用户已进入被监控状态
	loginServerEvent := h.newEvent(audit.EventTypeUserLoginToServer).(*audit.LoginToServerEvent)
	loginServerEvent.Timestamp = time.Now().UnixNano()
	loginServerEvent.ClientIP = h.userIP
	loginServerEvent.ServerIP = h.serverIP
	loginServerEvent.ID = uuid.New()
	loginServerEvent.Buffer = make(chan []byte, 1)

	// 创建文件事件存储器（单字符事件和其他分开存储，单字符事件用于后续回放）
	//fsNormal := audit.NewStore(audit.StoreFile, path.Join(SessionNormalEventRecordDir, strings.Join([]string{SessionNormalEventsStoreKeyPrefix, audit.EventTypeUserLoginToServer, h.serverIP, h.sessionID, loginServerEvent.ID.String()}, "_")))
	//
	//loginServerEvent.SetStore(&fsNormal)
	//go loginServerEvent.FlushBuffer(flushDone, audit.SessionEventBufferFlushInterval)
	// 更新登陆事件信息
	if err := loginServerEvent.WriteToBuffer(loginServerEvent); err != nil {
		common.Log.Errorf("Failed to write event to buffer")
	}
	return loginServerEvent.ID
}
