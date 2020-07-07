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
	ch chan []byte
}

func NewChanWriter() *ChanWriter {
	cw := ChanWriter{}
	cw.ch = make(chan []byte, 1024*8)
	return &cw
}

var KeyTabBytes = []byte{7}
var KeyBackSpaceBytes = []byte{8, 27, 91, 75}
var KeyEnterBytes = []byte{13, 10}
var KeySpaceBytes = []byte{108}

func isTabKey(p []byte) bool {
	return bytes.Compare(p, KeyTabBytes) == 0
}

func isBackSpace(p []byte) bool {
	return bytes.Compare(p, KeyBackSpaceBytes) == 0
}

func isSpace(p []byte) bool {
	return bytes.Compare(p, KeySpaceBytes) == 0
}

func (h *interactiveHandler) Write(p []byte) (n int, err error) {

	// 去除颜色控制字符
	p = bytes.ReplaceAll(p, vt100EscapeCodes.Green, []byte{})
	p = bytes.ReplaceAll(p, vt100EscapeCodes.Black, []byte{})
	p = bytes.ReplaceAll(p, vt100EscapeCodes.Blue, []byte{})
	p = bytes.ReplaceAll(p, vt100EscapeCodes.Cyan, []byte{})
	p = bytes.ReplaceAll(p, vt100EscapeCodes.Magenta, []byte{})
	p = bytes.ReplaceAll(p, vt100EscapeCodes.Red, []byte{})
	p = bytes.ReplaceAll(p, vt100EscapeCodes.White, []byte{})
	p = bytes.ReplaceAll(p, vt100EscapeCodes.Yellow, []byte{})
	p = bytes.ReplaceAll(p, vt100EscapeCodes.Reset, []byte{})

	// 当登入server，首先获取命令行提示符
	baa := bytes.Split(p, KeyEnterBytes)
	if len(baa) > 0 && len(baa[len(baa)-1]) > 0 {
		h.sessionPrompt = baa[len(baa)-1]
	}
	common.Log.Debugln("命令行提示符：", string(h.sessionPrompt))
	// 根据输入判断是否为命令字符（不以回车键字符开头,并且不包含命令提示符，则记为命令字符）
	if !bytes.HasPrefix(p, KeyEnterBytes) && !bytes.Contains(p, h.sessionPrompt) {
		common.Log.Println("command char: ", string(p))
		h.commandBuffer = append(h.commandBuffer, p...)
		h.kbEventWriter.ch <- p
		return 1, nil
	} else {
		// 根据输出判断是否为命令执行（如果输出以回车键字符开头，并且不等于回车键字符加命令行提示符，则记为一次命令执行）
		if len(h.commandBuffer) > 0 && bytes.HasPrefix(p, KeyEnterBytes) && bytes.Compare(p, append(KeyEnterBytes, h.sessionPrompt...)) != 0 {
			// 如果包含tab字符则需要去除
			if bytes.Contains(h.commandBuffer, KeyTabBytes) {
				h.commandBuffer = bytes.ReplaceAll(h.commandBuffer, KeyTabBytes, []byte{})
			}
			h.commandCh <- h.commandBuffer
			h.commandBuffer = make([]byte, 0)
			// 如果输出为回车键字符加命令行提示符，则为取消执行命令（执行了ctrl-c）
		} else if bytes.Compare(p, append(KeyEnterBytes, h.sessionPrompt...)) == 0 {
			h.commandBuffer = make([]byte, 0)
		}
		h.kbEventWriter.ch <- p
		return len(p), nil
	}
}

func (h *interactiveHandler) Close() (err error) {
	close(h.kbEventWriter.ch)
	return
}

func (h *interactiveHandler) Watch(event audit.IEvent, we chan bool) {
	defer close(event.GetBuffer())
	switch e := event.(type) {
	case *audit.KBEvent:
		for {
			select {
			case data := <-h.kbEventWriter.ch:
				e.Data = data
				e.Timestamp = time.Now().UnixNano()
				if err := e.WriteToBuffer(e); err != nil {
					common.Log.Errorf("couldn't write event to store")
				}
			case done := <-we:
				if done {
					common.Log.Infoln("exiting watcher")
					return
				}
			}
		}
	case *audit.ExecEvent:
		for {
			select {
			case data := <-h.commandCh:
				e.Command = string(data)
				common.Log.Printf("Full command string: %s", e.Command)
				e.Timestamp = time.Now().UnixNano()
				if err := e.WriteToBuffer(e); err != nil {
					common.Log.Errorf("couldn't write event to store")
				}
			case done := <-we:
				if done {
					common.Log.Infoln("exiting watcher")
					return
				}
			}
		}
	}
}
