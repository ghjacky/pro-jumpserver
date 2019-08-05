package assets

import (
	"fmt"
	"golang.org/x/crypto/ssh"
)

type ACommon struct {
	IP   string `json:"ip"`
	PORT int16  `json:"port"`
}

type IAsset interface {
	Connect() interface{}
	NewSession() interface{}
}

func NewAssetClient(as IAsset) (IAsset, error) {
	var err error
	switch v := as.(type) {
	case *ASSH:
		c := v.Connect()
		if c != nil {
			v.Client = c.(*ssh.Client)
		} else {
			err = fmt.Errorf("couldn't connect to remote server: %s:%d\n", v.IP, v.PORT)
		}
	}
	return as, err
}
