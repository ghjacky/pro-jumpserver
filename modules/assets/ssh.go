package assets

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"net"
	"zeus/common"
	"zeus/models"
)

type ASSH struct {
	ACommon
	USER   string      `json:"user"`
	PASS   string      `json:"pass"`
	ARGS   string      `json:"args"`
	Client *ssh.Client `json:"client"`
}

func (a *ASSH) Connect() (c interface{}) {
	sshConfig := &ssh.ClientConfig{
		User: a.USER,
		Auth: []ssh.AuthMethod{
			ssh.Password(a.PASS),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", net.JoinHostPort(a.IP, fmt.Sprintf("%d", a.PORT)), sshConfig)
	if err != nil {
		common.Log.Errorf("Couldn't connect remote host: %s", net.JoinHostPort(a.IP, fmt.Sprintf("%d", a.PORT)))
		return
	} else {
		c = client
	}
	return
}

func (a *ASSH) NewSession() (s interface{}) {
	var err error
	s, err = a.Client.NewSession()
	if err != nil {
		common.Log.Errorf("Couldn't connect to remote host: %s:%d using ssh", a.IP, a.PORT)
	}
	return
}

func FetchPermedAssets(user models.User, idc string) (ss []models.Server) {
	assets := user.FetchPermissionAssets(idc)
	for _, a := range assets {
		ss = append(ss, a.(*models.Asset).Server)
	}
	return
}
