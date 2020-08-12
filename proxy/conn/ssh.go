package conn

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"math/rand"
	"net"
	"time"
	"zeus/proxy/common"
)

func ConnectToSshServer(cw *SConnWrapper) (net.Conn, error) {
	sshClientC := &ssh.ClientConfig{
		User: cw.user,
		Auth: []ssh.AuthMethod{},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	if len(cw.pass) != 0 {
		sshClientC.Auth = append(sshClientC.Auth, ssh.Password(cw.pass))
	}
	if cw.keysig != nil {
		sshClientC.Auth = append(sshClientC.Auth, ssh.PublicKeys(cw.keysig))
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
	l, err := net.Listen("tcp", net.JoinHostPort("", fmt.Sprintf("%d", cw.prport)))
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
	return TunnelConnect(l, ctpconn, ptsconn)
}

func createHalfBehindTunnel(sc *ssh.Client, remote string) (net.Conn, error) {
	ptsconn, err := sc.Dial("tcp", remote)
	if err != nil {
		common.Log.Errorf("ssh Client dial error: %s", err.Error())
		return nil, err
	}
	return ptsconn, nil
}

func TunnelConnect(listener net.Listener, src, dest net.Conn) error {
	defer listener.Close()
	defer tunnelWipeout(src, dest)
	var tunnelDone = make(chan error, 2)
	go func() {
		if _, err := io.Copy(dest, src); err != nil {
			common.Log.Errorf("io copy error: %s", err.Error())
			tunnelDone <- fmt.Errorf("tunnel done")
			return
		} else {
			common.Log.Debugf("io copy done")
			tunnelDone <- fmt.Errorf("tunnel done")
		}
	}()
	go func() {
		if _, err := io.Copy(src, dest); err != nil {
			common.Log.Errorf("io copy error: %s", err.Error())
			tunnelDone <- fmt.Errorf("tunnel done")
			return
		} else {
			common.Log.Debugf("io copy done")
			tunnelDone <- fmt.Errorf("tunnel done")
		}
	}()
	<-tunnelDone
	common.Log.Infof("tunnel done")
	return nil
}

func isConnClosed(conn net.Conn) bool {
	one := make([]byte, 1)
	conn.SetReadDeadline(time.Now())
	if _, err := conn.Read(one); err == io.EOF {
		common.Log.Debugf("detect connection: %s <-> %s closed", conn.LocalAddr().String(), conn.RemoteAddr().String())
		return true
	} else if err == nil {
		conn.SetReadDeadline(time.Time{})
		return false
	} else if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
		common.Log.Infof("detect connection: %s <-> %s timeout error", conn.LocalAddr().String(), conn.RemoteAddr().String())
		return true
	}
	return false
}

func tunnelWipeout(src, dest net.Conn) {
	_ = src.Close()
	_ = dest.Close()
	src = nil
	dest = nil
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
