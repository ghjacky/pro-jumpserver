package jumpserver

import (
	"zeus/common"
	"zeus/utils"
)

func GetPubKey() []byte {
	var hostkey = utils.HostKey{
		Path:  "",
		Value: common.Config.PrivateKey,
	}
	sig, e := hostkey.Load()
	if e != nil {
		return []byte{}
	}
	kb := sig.PublicKey().Marshal()
	return kb
}
