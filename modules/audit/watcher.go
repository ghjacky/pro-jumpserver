package audit

import (
	"time"
	"zeus/common"
)

const (
	SessionEventBufferFlushInterval = 10 * time.Nanosecond
)

type IWatcher interface {
	Watch(event IEvent)
}

type ChanWriter struct {
	ch chan byte
}

func NewChanWriter() *ChanWriter {
	cw := ChanWriter{}
	cw.ch = make(chan byte, 1024)
	return &cw
}

func (cw *ChanWriter) Write(p []byte) (n int, err error) {
	for _, b := range p {
		cw.ch <- b
		n++
	}
	return
}

func (cw *ChanWriter) Close() (err error) {
	close(cw.ch)
	return
}

func (cw *ChanWriter) Watch(event IEvent, we chan bool) {
	defer close(event.GetBuffer())
	switch e := event.(type) {
	case *KBEvent:
		defer func() {
			if err := (*e.Store).(FileStore).Close(); err != nil {
				common.Log.Errorf("Failed to close file: %s", err.Error())
			}
		}()
		for {
			select {
			case data := <-cw.ch:
				e.Data = []byte{data}
				e.Timestamp = time.Now().UnixNano()
				if err := e.WriteToBuffer(e); err != nil {
					common.Log.Errorf("couldn't write event to file store")
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
