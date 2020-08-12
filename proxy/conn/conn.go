package conn

import (
	"encoding/json"
	"fmt"
	"github.com/gliderlabs/ssh"
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
	keysig ssh.Signer
	dip    string
	dport  uint16
}

const (
	SSHTIMEOUT = 15 * time.Second
)

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
	clientDataCh.SetTimeout(SSHTIMEOUT)
	go func() {
		utils.RecvFrom(c, clientDataCh)
	}()
	buffer, timeout := clientDataCh.ReadWithTimeout()
	if timeout {
		common.Log.Errorln("read from client timeout")
		SendBack(c, "", "socket read error", 0)
		return
	}
	dd := []byte(utils.Dec(string(buffer)))
	var req = protocol.NewMessage()
	if err := json.Unmarshal(dd, req); err != nil {
		SendBack(c, cw.ppip, fmt.Sprintf("data unmarshal error: %s", err.Error()), 0)
		return
	}
	if ok, err := req.Valid(); ok && req.IsReq() {
		cw.dip = req.GetDip()
		cw.dport = req.GetDPort()
		cw.user = req.GetUser()
		cw.pass = req.GetPass()
		cw.keysig = req.GetKeySig()
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
	if len(err) != 0 {
		common.Log.Warnf("error to be send back: %s", err)
	}
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
	e = utils.SendTo(c, data)
	if e != nil {
		common.Log.Errorf("data send back error: %s", e.Error())
		return
	}
}
