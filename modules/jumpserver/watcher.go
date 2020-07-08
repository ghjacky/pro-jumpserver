package jumpserver

import (
	"bytes"
	"time"
	"zeus/common"
	"zeus/modules/audit"
)

const (
	SessionEventBufferFlushInterval = 10 * time.Nanosecond
)

type IWatcher interface {
	Watch(event audit.IEvent)
}

type ChanWriter struct {
	ch        chan []byte
	tabFlag   bool
	cbHandler *interactiveHandler
}

type LineChanWriter struct {
	ch        chan []byte
	bf        []byte
	cbHandler *interactiveHandler
}

func NewChanWriter(h *interactiveHandler) *ChanWriter {
	cw := ChanWriter{}
	cw.ch = make(chan []byte, 1024*8)
	cw.cbHandler = h
	return &cw
}

func NewLineChanWriter(h *interactiveHandler) *LineChanWriter {
	lcw := LineChanWriter{}
	lcw.ch = make(chan []byte, 1024)
	lcw.bf = []byte{}
	lcw.cbHandler = h
	return &lcw
}

const (
	KeyTabOutput       int32 = 7
	KeySlashOutput           = 47
	KeyPrintableOutput       = 32
)

func (cw *ChanWriter) Write(p []byte) (n int, err error) {
	key, rest := bytesToKey(p, false)
	switch key {
	// 自动补全命令
	case KeyTabOutput:
		if cw.tabFlag {
			cw.cbHandler.execEventWriter.Write(rest)
			cw.tabFlag = !cw.tabFlag
		}
	default:
		// 自动补全路径
		if cw.tabFlag && (bytes.HasSuffix(p, []byte{KeySlashOutput}) || bytes.HasSuffix(p, []byte{KeyPrintableOutput})) {
			cw.cbHandler.execEventWriter.Write(p)
			cw.tabFlag = !cw.tabFlag
		}
	}
	cw.ch <- p
	return len(p), nil
}

func (cw *ChanWriter) Close() (err error) {
	close(cw.ch)
	return
}

var (
	KeyAltRightInput     = []byte{27, 27, 91, 67}
	KeyAltLeftInput      = []byte{27, 27, 91, 68}
	KeyCommandLeftInput  = []byte{27, 98}
	KeyCommandRightInput = []byte{27, 102}
)

func (lcw *LineChanWriter) setCursXToPos(pos int) {
	lcw.cbHandler.term.moveCursorToPos(pos)
	lcw.cbHandler.term.pos = pos
}
func (lcw *LineChanWriter) SetBF(bs []byte) {
	lcw.bf = bs
	var line []rune
	for _, b := range bs {
		line = append(line, rune(b))
	}
	lcw.cbHandler.term.line = line
}

// 此处write方法需要判断命令输入，以执行命令为单位进行byte写入
func (lcw *LineChanWriter) Write(p []byte) (n int, err error) {
	nr := len(p)
	key, _ := bytesToKey(p, false)
	switch key {
	case keyEnter:
		if len(lcw.bf) > 0 {
			lcw.cbHandler.term.history.Add(string(lcw.bf))
			lcw.cbHandler.term.historyIndex = -1
			lcw.ch <- lcw.bf
			lcw.SetBF([]byte{})
		}
		lcw.setCursXToPos(0)
	case keyCtrlC, keyCtrlU, keyCtrlD, keyDeleteLine:
		lcw.SetBF([]byte{})
		lcw.setCursXToPos(0)
		lcw.cbHandler.term.historyIndex = -1
	case keyBackspace:
		if len(lcw.bf) > 1 && lcw.cbHandler.term.cursorX > 0 {
			lcw.SetBF(append(lcw.bf[:lcw.cbHandler.term.cursorX-1], lcw.bf[lcw.cbHandler.term.cursorX:]...))
			lcw.setCursXToPos(lcw.cbHandler.term.cursorX - 1)
		}
	case keyTab:
		lcw.cbHandler.kbEventWriter.tabFlag = true
	case keyLeft:
		if lcw.cbHandler.term.cursorX >= 1 {
			lcw.setCursXToPos(lcw.cbHandler.term.cursorX - 1)
		}
	case keyRight:
		if lcw.cbHandler.term.cursorX <= len(lcw.bf)-1 {
			lcw.setCursXToPos(lcw.cbHandler.term.cursorX + 1)
		}
	case keyUp:
		if lcw.cbHandler.term.historyIndex == -1 {
			lcw.cbHandler.term.historyPending = string(lcw.bf)
		}
		if lcw.cbHandler.term.historyIndex < len(lcw.cbHandler.term.history.entries)-1 {
			lcw.cbHandler.term.historyIndex++
		}
		entry, ok := lcw.cbHandler.term.history.NthPreviousEntry(lcw.cbHandler.term.historyIndex)
		if ok {
			lcw.SetBF([]byte(entry))
			lcw.setCursXToPos(len(entry))
		}
	case keyDown:
		if lcw.cbHandler.term.historyIndex >= 0 {
			lcw.cbHandler.term.historyIndex--
		}
		if lcw.cbHandler.term.historyIndex != -1 {
			entry, ok := lcw.cbHandler.term.history.NthPreviousEntry(lcw.cbHandler.term.historyIndex)
			if ok {
				lcw.setCursXToPos(len(entry))
				lcw.SetBF([]byte(entry))
			}
		} else {
			lcw.SetBF([]byte(lcw.cbHandler.term.historyPending))
		}
	case keyDeleteWord:
		beforeX := lcw.cbHandler.term.cursorX
		n := lcw.cbHandler.term.countToLeftWord()
		lcw.cbHandler.term.eraseNPreviousChars(n)
		afterX := lcw.cbHandler.term.cursorX
		lcw.SetBF(append(lcw.bf[:afterX], lcw.bf[beforeX:]...))
	default:
		// 为什么要转换为string？因为string是值类型，slice是引用类型，嵌套append为使内层变量在append完之后保持不变需要使用完全复制的变量，
		// 使用string类型即变相的实现了slice的完全复制
		after := string(lcw.bf[lcw.cbHandler.term.cursorX:])
		before := string(lcw.bf[:lcw.cbHandler.term.cursorX])
		lcw.SetBF(append(append([]byte(before), p...), after...))
		lcw.setCursXToPos(lcw.cbHandler.term.cursorX + len(p))
	}
	return nr, nil
}

func (lcw *LineChanWriter) Close() (err error) {
	close(lcw.ch)
	return
}

func (h *interactiveHandler) Watch(event audit.IEvent, we chan bool) {
	defer close(event.GetBuffer())
	switch e := event.(type) {
	case *audit.KBEvent:
		common.Log.Infof("Starting record kb event")
		for {
			select {
			case data := <-h.kbEventWriter.ch:
				e.Data = data
				e.Timestamp = time.Now().UnixNano()
				if err := e.WriteToBuffer(e); err != nil {
					common.Log.Errorf("couldn't write kb event to store")
				}
			case done := <-we:
				if done {
					common.Log.Infoln("exiting kb watcher")
					return
				}
			}
		}
	case *audit.ExecEvent:
		common.Log.Infof("Starting record exec event")
		for {
			select {
			case data := <-h.execEventWriter.ch:
				e.Command = string(data)
				common.Log.Printf("Full command string: %s", e.Command)
				e.Timestamp = time.Now().UnixNano()
				if err := e.WriteToBuffer(e); err != nil {
					common.Log.Errorf("couldn't write exec event to store")
				}
			case done := <-we:
				if done {
					common.Log.Infoln("exiting exec watcher")
					return
				}
			}
		}
	}
}
