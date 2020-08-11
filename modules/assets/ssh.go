package assets

import (
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh"
	"net"
	"time"
	"zeus/common"
	"zeus/models"
	"zeus/proxy/protocol"
	"zeus/utils"
)

type ASSH struct {
	ACommon
	USER   string      `json:"user"`
	PASS   string      `json:"pass"`
	ARGS   string      `json:"args"`
	KEYSIG ssh.Signer  `json:"key_sig"`
	Client *ssh.Client `json:"client"`
}

const SSHTIMEOUT = 15 * time.Second

// Connect 远端主机ssh连接；首先判断资产所属IDC，进一步判断通往此IDC是否设置有相应的代理，如果有设置代理，则需要连接代理，
// 并进而通过代理将ssh连接转发到远端主机。
func (a *ASSH) Connect() (c interface{}) {
	scc := &ssh.ClientConfig{
		User:            a.USER,
		Auth:            []ssh.AuthMethod{},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         SSHTIMEOUT,
	}
	if len(a.PASS) != 0 {
		scc.Auth = append(scc.Auth, ssh.Password(a.PASS))
	}
	if a.KEYSIG != nil {
		scc.Auth = append(scc.Auth, ssh.PublicKeys(a.KEYSIG))
	}
	var ip = a.IP
	var port = a.PORT
	var idc = models.SIDC{Name: a.IDC}
	// 如果需要代理，连接代理控制端口发送必要信息，并通过返回的信息连接映射端口
	if need, err := idc.NeedProxy(); need && err == nil {
		ppip := fmt.Sprintf("%d.%d.%d.%d", idc.Proxy.PPIP[0], idc.Proxy.PPIP[1], idc.Proxy.PPIP[2], idc.Proxy.PPIP[3])
		pip := fmt.Sprintf("%d.%d.%d.%d", idc.Proxy.PIP[0], idc.Proxy.PIP[1], idc.Proxy.PIP[2], idc.Proxy.PIP[3])
		pport := idc.Proxy.PPORT
		req := new(protocol.SProtocol)
		req.SetT(0)
		req.SetDip(a.IP)
		req.SetDPort(a.PORT)
		req.SetPPip(ppip)
		req.SetPip(pip)
		req.SetPPort(pport)
		req.SetUser(a.USER)
		req.SetPass(a.PASS)
		req.SetKeySig(a.KEYSIG)
		common.Log.Debugf("proxy request: %#v", *req)
		var respChan = utils.NewTimeoutChan(make(chan []byte, 0))
		respChan.SetTimeout(15 * time.Second)
		proxy := connectToProxy(ppip, pport)
		if proxy == nil {
			common.Log.Errorf("connect to proxy error")
			return
		}
		defer proxy.Close()
		resp, err := a.useProxy(*req, proxy, respChan)
		if err != nil {
			common.Log.Errorf("use proxy error: %s", err.Error())
			return
		}
		ip = resp.GetPPip()
		port = resp.GetPRPort()
	} else if need && err != nil {
		common.Log.Errorf("use proxy error: %s", err.Error())
		return
	}
	// 连接
	if sc, err := connectToSSH(ip, port, scc); err != nil {
		common.Log.Errorf("Couldn't connect remote host: %s", net.JoinHostPort(ip, fmt.Sprintf("%d", port)))
		return
	} else {
		c = sc
		return
	}
}

func (a *ASSH) NewSession() (s interface{}) {
	var err error
	s, err = a.Client.NewSession()
	if err != nil {
		common.Log.Errorf("Couldn't connect to remote host: %s:%d using ssh", a.IP, a.PORT)
	}
	return
}

func (a *ASSH) useProxy(req protocol.SProtocol, proxy net.Conn, respChan *utils.TimeoutChan) (*protocol.SProtocol, error) {
	go recvFromProxy(proxy, respChan)
	if err := sendToProxy(proxy, req); err != nil {
		return nil, err
	}
	buf, timeout := respChan.ReadWithTimeout()
	if timeout {
		return nil, fmt.Errorf("read from proxy timeout")
	}
	common.Log.Infof("client received data from proxy: %s", string(buf))
	var resp = protocol.NewMessage()
	if err := json.Unmarshal(buf, resp); err != nil {
		return resp, err
	}
	if err := resp.GetErr(); err != nil {
		common.Log.Errorf("error from response: %s", err.Error())
		return resp, err
	}
	return resp, nil
}

func connectToProxy(ip string, port uint16) net.Conn {
	proxy, err := net.Dial("tcp", net.JoinHostPort(ip, fmt.Sprintf("%d", port)))
	if err != nil {
		common.Log.Errorf("proxy(%s) connect failed: %s", net.JoinHostPort(ip, fmt.Sprintf("%d", port)), err.Error())
		return nil
	}
	return proxy
}

func sendToProxy(c net.Conn, req protocol.SProtocol) error {
	d, err := json.Marshal(req)
	ed := []byte(utils.Enc(string(d)))
	if err != nil {
		common.Log.Errorf("req marshal error: %s", err.Error())
		return err
	}
	if _, err := c.Write(ed); err != nil {
		common.Log.Errorf("failed to send req to proxy: %s", err.Error())
		return err
	}
	common.Log.Debugln("client send data to proxy successfully")
	return nil
}

func recvFromProxy(c net.Conn, respChan *utils.TimeoutChan) {
	var buf = make([]byte, 0)
	var tbf = make([]byte, 256)
	for {
		n, err := c.Read(tbf)
		if err != nil {
			common.Log.Warnf("proxy may be disconnected: %s", err.Error())
			break
		}
		buf = append(buf, tbf[:n]...)
		if n < 256 {
			if timeout := respChan.WriteWithTimeout(buf); timeout {
				return
			}
			buf = make([]byte, 0)
			tbf = make([]byte, 256)
		}
	}
}

func connectToSSH(ip string, port uint16, scc *ssh.ClientConfig) (*ssh.Client, error) {
	client, err := ssh.Dial("tcp", net.JoinHostPort(ip, fmt.Sprintf("%d", port)), scc)
	if err != nil {
		return nil, err
	} else {
		return client, nil
	}
}
