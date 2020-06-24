package audit

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
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
	GetStore() *IStore
	Marshal(IEvent) []byte
	WriteToBuffer(IEvent) error
	FlushBuffer(chan interface{}, time.Duration)
	GetBuffer() chan []byte
}

// 定义事件基本类型
type Event struct {
	ID        uuid.UUID
	SessionID string  // JumpServer session id
	Err       string  // 事件错误message，为空则为成功事件
	Type      string  // 事件类型
	User      string  // 触发事件的用户
	Timestamp int64   // 事件发生时间戳
	ClientIP  string  // 登陆客户端地址
	ServerIP  string  // 登陆服务端地址
	Store     *IStore // 事件存储器接口
	Buffer    chan []byte
}

func (e *Event) GetBuffer() chan []byte {
	return e.Buffer
}
func (*Event) Marshal(e IEvent) (data []byte) {
	var err error
	var me models.Event
	switch v := e.(type) {
	case *LoginToJpsEvent:
		me = models.Event{
			SessionID: v.SessionID,
			Type:      v.Type,
			Err:       v.Err,
			User:      v.User,
			Timestamp: v.Timestamp,
			ClientIP:  v.ClientIP,
			ServerIP:  v.ServerIP,
		}
		me.ID = v.ID
		// 登陆事件相关部分信息写入mysql
		common.Mysql.Create(&me)
	case *LoginToServerEvent:
		me = models.Event{
			SessionID: v.SessionID,
			Type:      v.Type,
			Err:       v.Err,
			User:      v.User,
			Timestamp: v.Timestamp,
			ClientIP:  v.ClientIP,
			ServerIP:  v.ServerIP,
		}
		me.ID = v.ID
		// 登陆事件相关部分信息入库（db）
		common.Mysql.Create(&me)
	case *ExecEvent:
		me = models.Event{
			SessionID: v.SessionID,
			Type:      v.Type,
			Err:       v.Err,
			User:      v.User,
			Timestamp: v.Timestamp,
			ClientIP:  v.ClientIP,
			ServerIP:  v.ServerIP,
			Bin:       v.Bin,
			Command:   v.Command,
		}
	case *FileEvent:
		me = models.Event{
			SessionID: v.SessionID,
			Type:      v.Type,
			Err:       v.Err,
			User:      v.User,
			Timestamp: v.Timestamp,
			ClientIP:  v.ClientIP,
			ServerIP:  v.ServerIP,
			SrcFile:   v.SrcFile,
			DestFile:  v.DestFile,
		}
	case *KBEvent:
		me = models.Event{
			SessionID: v.SessionID,
			Type:      v.Type,
			Err:       v.Err,
			User:      v.User,
			Timestamp: v.Timestamp,
			ClientIP:  v.ClientIP,
			ServerIP:  v.ServerIP,
			Data:      v.Data,
		}
	}
	// 仅序列化必要的字段，以最大限度的减小事件占用的存储大小
	data, err = json.Marshal(map[string]interface{}{"timestamp": me.Timestamp, "data": me.Data})
	if err != nil {
		common.Log.Errorf("Couldn't marshal event to byte array")
	}
	return
}

func (e *Event) SetStore(store *IStore) {
	e.Store = store
	return
}

func (e *Event) GetStore() (store *IStore) {
	store = e.Store
	return
}

func (e *Event) FlushBuffer(done chan interface{}, interval time.Duration) {
	defer func() {
		if err := (*e.Store).Close(); err != nil {
			common.Log.Errorln("Couldn't close events store")
		}
	}()
	for {
		//time.Sleep(interval)
		select {
		case event := <-e.Buffer:
			if event != nil {
				_, err := (*e.Store).Write(append(event, []byte("\n")...))
				if err != nil {
					return
				}
			} else {
				return
			}
		case <-done:
			return
		}
	}
}

// 将事件写入buffer，实现timeout
func (*Event) WriteToBuffer(e IEvent) (err error) {
	wg := sync.WaitGroup{}
	lock := sync.Mutex{}
	isSuc := make(chan bool, 0)
	go func() {
		e.GetBuffer() <- e.Marshal(e)
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
	wg.Wait()
	return
}

// 定义登陆事件类型
type LoginToJpsEvent struct {
	Event
}
type LoginToServerEvent struct {
	Event
}

// 根据用户名、时间段、Server地址查询登陆事件
func (le *LoginToServerEvent) Search(args ...interface{}) (es []IEvent) {

	return
}

// 根据用户名、时间段、Server地址查询登陆事件
func (le *LoginToJpsEvent) Search(args ...interface{}) (es []IEvent) {

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
