package audit

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"zeus/common"
	"zeus/models"
)

const (
	// 事件类型常量
	EventTypeUserLoginToJPS          = "JPSLogin"
	EventTypeUserLoginToServer       = "AssetsLogin"
	EventTypeUserExecCommandOnJPS    = "ExecOnJPS"
	EventTypeUserExecCommandOnServer = "ExecOnServer"
	EventTypeUserUploadFile          = "UploadFile"
	EventTypeUserDownloadFile        = "DownloadFile"
	EventTypeKeyBoardPress           = "KeyBoardPress"
	EventWriteTimeout                = 1 * time.Second
)

// 定义事件接口
type IEvent interface {
	Search(...interface{}) []IEvent
	SetStore(*IStore)
	Marshal(IEvent) []byte
	WriteToBuffer(IEvent, chan []byte) error
	FlushBuffer(chan []byte, time.Duration)
}

// 定义事件基本类型
type Event struct {
	SessionID    string  // JumpServer session id
	SubSessionID string  // 登陆资产session id
	Err          string  // 事件错误message，为空则为成功事件
	Type         string  // 事件类型
	User         string  // 触发事件的用户
	Timestamp    int64   // 事件发生时间戳
	ClientIP     string  // 登陆客户端地址
	ServerIP     string  // 登陆服务端地址
	Store        *IStore // 事件存储器接口
}

func (*Event) Marshal(e IEvent) (data []byte) {
	var err error
	var me models.Event
	switch v := e.(type) {
	case *LoginEvent:
		me = models.Event{
			v.SessionID,
			v.SubSessionID,
			v.Type,
			v.Err,
			v.User,
			v.Timestamp,
			v.ClientIP,
			v.ServerIP,
			"",
			"",
			"",
			"",
			[]byte{},
		}
	case *ExecEvent:
		me = models.Event{
			v.SessionID,
			v.SubSessionID,
			v.Type,
			v.Err,
			v.User,
			v.Timestamp,
			v.ClientIP,
			v.ServerIP,
			"",
			"",
			v.Bin,
			v.Command,
			[]byte{},
		}
	case *FileEvent:
		me = models.Event{
			v.SessionID,
			v.SubSessionID,
			v.Type,
			v.Err,
			v.User,
			v.Timestamp,
			v.ClientIP,
			v.ServerIP,
			v.SrcFile,
			v.DestFile,
			"",
			"",
			[]byte{},
		}
	case *KBEvent:
		me = models.Event{
			v.SessionID,
			v.SubSessionID,
			v.Type,
			v.Err,
			v.User,
			v.Timestamp,
			v.ClientIP,
			v.ServerIP,
			"",
			"",
			"",
			"",
			v.Data,
		}
	}
	data, err = json.Marshal(me)
	if err != nil {
		common.Log.Errorf("Couldn't marshal event to byte array")
	}
	return
}

func (e *Event) SetStore(store *IStore) {
	e.Store = store
	return
}

func (e *Event) FlushBuffer(b chan []byte, interval time.Duration) {
	for {
		time.Sleep(interval)
		select {
		case event := <-b:
			if len(event) > 0 {
				fmt.Println("got event, starting to flush event to store")
				_, _ = (*e.Store).Write(event)
			}
		default:
			continue
		}
	}
}

// 将事件写入buffer，实现timeout
func (*Event) WriteToBuffer(e IEvent, b chan []byte) (err error) {
	fmt.Println("strting write to buffer")
	wg := sync.WaitGroup{}
	lock := sync.Mutex{}
	isSuc := make(chan bool, 0)
	go func() {
		b <- e.Marshal(e)
		lock.Lock()
		isSuc <- true
		lock.Unlock()
	}()
	go func() {
		time.Sleep(EventWriteTimeout)
		lock.Lock()
		isSuc <- false
		lock.Unlock()
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case b := <-isSuc:
			if !b {
				err = fmt.Errorf("couldn't write event to buffer")
			}
		}
	}()
	fmt.Printf("res of writing event to buffer: %s", err)
	wg.Wait()
	return
}

// 定义登陆事件类型
type LoginEvent struct {
	Event
}

// 根据用户名、时间段、Server地址查询登陆事件
func (le *LoginEvent) Search(args ...interface{}) (es []IEvent) {

	return
}

// 定义用户执行命令事件类型
type ExecEvent struct {
	Event
	Bin     string // 命令名字
	Command string // 完整命令字串
}

func (ee *ExecEvent) Search(args ...interface{}) (es []IEvent) {

	return
}

// 定义文件相关事件类型
type FileEvent struct {
	Event
	SrcFile  string // 源文件路径
	DestFile string // 目标文件路径
}

func (fe *FileEvent) Search(args ...interface{}) (es []IEvent) {

	return
}

// 按键事件类型(通过os.stdout记录，所以只会记录可见的stdout事件)
type KBEvent struct {
	Event
	Data []byte // 字符
}

func (ke *KBEvent) Search(args ...interface{}) (es []IEvent) {
	return
}

// 生成事件
func NewEvent(t string) (e IEvent) {
	switch t {
	case EventTypeUserLoginToJPS:
		e = &LoginEvent{Event{Type: EventTypeUserLoginToJPS}}
	}
	return
}
