package assets

import "golang.org/x/crypto/ssh"

type ACommon struct {
	IP   string `json:"ip"`
	PORT int16  `json:"port"`
}

type IAsset interface {
	Connect() interface{}
	NewSession() interface{}
}

func NewAssetClient(as IAsset) IAsset {
	switch v := as.(type) {
	case *ASSH:
		c := v.Connect()
		if c != nil {
			v.Client = c.(*ssh.Client)
		}
	}
	return as
}
