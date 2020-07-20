package conn

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"math/rand"
	"net"
	"zeus/common"
)

func ConnectToSshServer(cw *SConnWrapper) (net.Conn, error) {
	sshClientC := &ssh.ClientConfig{
		User: cw.user,
		Auth: []ssh.AuthMethod{
			ssh.Password(cw.pass),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	sshClient, err := ssh.Dial("tcp", net.JoinHostPort(cw.dip, fmt.Sprintf("%d", cw.dport)), sshClientC)
	if err != nil {
		common.Log.Errorf("ssh dial error: %s", err.Error())
		return nil, err
	}
	randPort := randomValidPort()
	cw.prport = randPort
	return createHalfBehindTunnel(sshClient, net.JoinHostPort(cw.dip, fmt.Sprintf("%d", cw.dport)))
}

func HalfAheadTunnelListenOn(cw *SConnWrapper, ptsconn net.Conn) error {
	l, err := net.Listen("tcp", net.JoinHostPort(cw.ppip, fmt.Sprintf("%d", cw.prport)))
	if err != nil {
		common.Log.Errorf("listener run error: %s", err.Error())
		return err
	}
	SendBack(cw.conn, cw.ppip, "", cw.prport)
	ctpconn, err := l.Accept()
	if err != nil {
		common.Log.Errorf("listener connect error: %s", err.Error())
		return err
	}
	return TunnelConnect(ctpconn, ptsconn)
}

func createHalfBehindTunnel(sc *ssh.Client, remote string) (net.Conn, error) {
	ptsconn, err := sc.Dial("tcp", remote)
	if err != nil {
		common.Log.Errorf("ssh Client dial error: %s", err.Error())
		return nil, err
	}
	return ptsconn, nil
}

func TunnelConnect(src, dest net.Conn) error {
	defer tunnelWipeout(src, dest)
	var tunnelDone = make(chan error, 1)
	go func() {
		if _, err := io.Copy(dest, src); err != nil {
			common.Log.Errorf("io copy error: %s", err.Error())
			tunnelDone <- fmt.Errorf("tunnel done")
			return
		}
	}()
	go func() {
		if _, err := io.Copy(src, dest); err != nil {
			common.Log.Errorf("io copy error: %s", err.Error())
			tunnelDone <- fmt.Errorf("tunnel done")
			return
		}
	}()
	<-tunnelDone
	return nil
}

func tunnelWipeout(src, dest net.Conn) {
	_ = src.Close()
	_ = dest.Close()
}

func randomValidPort() uint16 {
	max := 65535
	min := 20000
	for {
		port := uint16(rand.Intn(max-min) + min)
		if checkIfPortAvailable(port) {
			return port
		}
	}
}

func checkIfPortAvailable(port uint16) bool {
	l, err := net.Listen("tcp", net.JoinHostPort("0.0.0.0", fmt.Sprintf("%d", port)))
	if err != nil {
		return false
	}
	err = l.Close()
	if err != nil {
		return false
	}
	return true
}
