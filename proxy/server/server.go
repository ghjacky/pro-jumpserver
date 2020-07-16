package server

import (
	"net"
	"time"
	"zeus/proxy/common"
	"zeus/proxy/conn"
)

const (
	ServeThread = 10
	ConnTimeout = 5 * time.Second
)

var ServeDone = make(chan uint8, 0)

func ProxyServerRun() {
	controlServerRun(common.Config.Listen)
	common.Log.Infof("服务正运行于：%s ...", common.Config.Listen)
	waitServerDone(ServeDone)
}

func controlServerRun(listen string) {
	server, err := net.Listen("tcp4", listen)
	if err != nil {
		common.Log.Fatalf("服务监听失败： %s", err.Error())
	}
	for i := ServeThread; i > 0; i-- {
		go func() {
			for {
				c, err := server.Accept()
				if err != nil {
					common.Log.Errorf("连接建立失败： %s", err.Error())
					return
				}
				_ = c.SetDeadline(time.Now().Add(ConnTimeout))
				go conn.HandleClientConn(c)
			}
		}()
	}
}

func waitServerDone(done <-chan uint8) {
	<-done
}
