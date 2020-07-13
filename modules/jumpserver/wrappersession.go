package jumpserver

import (
	"github.com/gliderlabs/ssh"
	"io"
	"net"
	"sync"
	"zeus/common"
)

type WrapperSession struct {
	tabKCount uint
	Sess      ssh.Session
	inWriter  io.WriteCloser
	outReader io.ReadCloser
	mux       *sync.RWMutex
}

func (w *WrapperSession) initial() {
	w.initReadPip()
	go w.readLoop()
}

func (w *WrapperSession) readLoop() {
	defer func() {
		common.Log.Debug("session loop break")
		//JpsFlushDone <- 1
	}()
	buf := make([]byte, 1024*8)
	for {
		nr, err := w.Sess.Read(buf)

		if nr > 0 {
			w.mux.RLock()
			_, _ = w.inWriter.Write(buf[:nr])
			w.mux.RUnlock()
		}
		if err != nil {
			break
		}
	}
	_ = w.inWriter.Close()
}

func (w *WrapperSession) Read(p []byte) (int, error) {
	w.mux.RLock()
	defer w.mux.RUnlock()
	n, e := w.outReader.Read(p)
	// 监控tab按键，使用回调函数自动补全或展现所有相关主机
	key, _ := bytesToKey(p, false)
	switch key {
	case keyTab:
		w.tabKCount++
		if w.tabKCount == 2 {
			p[0] = keyEnter
			w.tabKCount = 0
		}
	default:
		w.tabKCount = 0
	}
	return n, e
}

func (w *WrapperSession) Close() error {
	var err error
	err = w.inWriter.Close()
	w.initReadPip()
	return err
}

// 此处为向ssh.session写入数据
func (w *WrapperSession) Write(p []byte) (int, error) {
	return w.Sess.Write(p)
}

func (w *WrapperSession) initReadPip() {
	w.mux.Lock()
	defer w.mux.Unlock()
	w.outReader, w.inWriter = io.Pipe()
}

func (w *WrapperSession) Protocol() string {
	return "ssh"
}

func (w *WrapperSession) User() string {
	return w.Sess.User()
}

func (w *WrapperSession) WinCh() (winch <-chan ssh.Window) {
	_, winch, ok := w.Sess.Pty()
	if ok {
		return
	}
	return nil
}

func (w *WrapperSession) LoginFrom() string {
	return "ST"
}

func (w *WrapperSession) RemoteAddr() string {
	host, _, _ := net.SplitHostPort(w.Sess.RemoteAddr().String())
	return host
}

func (w *WrapperSession) Pty() ssh.Pty {
	pty, _, _ := w.Sess.Pty()
	return pty
}

func NewWrapperSession(sess ssh.Session) *WrapperSession {
	w := &WrapperSession{
		Sess: sess,
		mux:  new(sync.RWMutex),
	}
	w.initial()
	return w
}
