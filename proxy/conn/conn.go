package conn

import (
	"encoding/json"
	"fmt"
	"net"
	"zeus/proxy/common"
	"zeus/proxy/protocol"
)

type SConnWrapper struct {
	conn   net.Conn
	pip    string
	prport uint16 // 此处为random port
	user   string
	pass   string
	dip    string
	dport  uint16
}

// HandleSrcConn 客户端连接处理函数， 接受数据包，解析数据包，获取ssh client相关信息，然后connect ssh server，
// 调用相关函数进行连接双向绑定
var CPool = NewConnPool()

func newConnWrapper(c net.Conn) *SConnWrapper {
	return &SConnWrapper{
		conn: c,
	}
}

func HandleClientConn(c net.Conn) {
	CPool.Put(c)
	defer CPool.WipeOut(c)
	cw := newConnWrapper(c)
	buffer := make([]byte, 1024)
	if _, err := c.Read(buffer); err != nil {
		SendBack(c, fmt.Sprintf("socket read error: %s", err.Error()), 0)
		return
	}
	var req = protocol.NewMessage()
	if err := json.Unmarshal(buffer, &req); err != nil {
		SendBack(c, fmt.Sprintf("data unmarshal error: %s", err.Error()), 0)
		return
	}
	if ok, err := req.Valid(); ok && req.IsReq() {
		cw.dip = req.GetDip()
		cw.dport = req.GetDPort()
		cw.user = req.GetUser()
		cw.pass = req.GetPass()
		cw.pip = req.GetPip()
		ptsconn, e := ConnectToSshServer(cw)
		if e != nil {
			SendBack(c, fmt.Sprintf("failed to connect to server: %s", e.Error()), 0)
			return
		}
		if e := HalfAheadTunnelListenOn(cw, ptsconn); e != nil {
			SendBack(c, fmt.Sprintf("tunnel listener run error: %s", e.Error()), 0)
			return
		}
	} else if err != nil {
		SendBack(c, fmt.Sprintf("request message error: %s", err.Error()), 0)
		return
	}
}

func SendBack(c net.Conn, err string, prport uint16) {
	resp := protocol.NewMessage()
	resp.SetT(1)
	resp.SetErr(err)
	resp.SetPRPort(prport)
	data, e := json.Marshal(resp)
	if e != nil {
		common.Log.Errorf("response data marshal eror: %s", e.Error())
		return
	}
	_, e = c.Write(data)
	if e != nil {
		common.Log.Errorf("data send back error: %s", e.Error())
		return
	}
}
