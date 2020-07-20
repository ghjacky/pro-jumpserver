package conn

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
	"zeus/proxy/common"
	"zeus/proxy/protocol"
	"zeus/utils"
)

type SConnWrapper struct {
	conn   net.Conn
	ppip   string // public ip of proxy
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
	var clientDataCh = utils.NewTimeoutChan(make(chan []byte, 0))
	clientDataCh.SetTimeout(15 * time.Second)
	go func() {
		buf := make([]byte, 0)
		tbf := make([]byte, 256)
		for {
			n, err := c.Read(tbf)
			if err != nil {
				SendBack(c, cw.ppip, fmt.Sprintf("socket read error: %s", err.Error()), 0)
				break
			}
			buf = append(buf, tbf[:n]...)
			if n < 256 {
				if timeout := clientDataCh.WriteWithTimeout(buf); timeout {
					common.Log.Errorln("sent to client timeout")
					return
				}
				buf = make([]byte, 0)
				tbf = make([]byte, 256)
			}
		}
	}()
	buffer, timeout := clientDataCh.ReadWithTimeout()
	if timeout {
		common.Log.Errorln("read from client timeout")
		return
	}
	common.Log.Infof("proxy received data: %s", string(buffer))
	var req = protocol.NewMessage()
	if err := json.Unmarshal(buffer, req); err != nil {
		SendBack(c, cw.ppip, fmt.Sprintf("data unmarshal error: %s", err.Error()), 0)
		return
	}
	if ok, err := req.Valid(); ok && req.IsReq() {
		cw.dip = req.GetDip()
		cw.dport = req.GetDPort()
		cw.user = req.GetUser()
		cw.pass = req.GetPass()
		cw.ppip = req.GetPPip()
		ptsconn, e := ConnectToSshServer(cw)
		if e != nil {
			SendBack(c, cw.ppip, fmt.Sprintf("failed to connect to server: %s", e.Error()), 0)
			return
		}
		if e := HalfAheadTunnelListenOn(cw, ptsconn); e != nil {
			SendBack(c, cw.ppip, fmt.Sprintf("tunnel listener run error: %s", e.Error()), 0)
			return
		}
	} else if err != nil {
		SendBack(c, cw.ppip, fmt.Sprintf("request message error: %s", err.Error()), 0)
		return
	}
}

func SendBack(c net.Conn, ppip, err string, prport uint16) {
	resp := protocol.NewMessage()
	resp.SetT(1)
	if err := resp.SetPPip(ppip); err != nil {
		h, _, _ := net.SplitHostPort(c.LocalAddr().String())
		_ = resp.SetPPip(h)
	}
	resp.SetErr(err)
	resp.SetPRPort(prport)
	data, e := json.Marshal(*resp)
	if e != nil {
		common.Log.Errorf("response data marshal eror: %s", e.Error())
		return
	}
	common.Log.Debugf("send response to client: %s", string(data))
	_, e = c.Write(data)
	if e != nil {
		common.Log.Errorf("data send back error: %s", e.Error())
		return
	}
}
