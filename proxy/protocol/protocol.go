package protocol

import (
	"fmt"
	"github.com/gliderlabs/ssh"
	"os"
	"strconv"
	"strings"
	"zeus/common"
	"zeus/utils"
)

type SProtocol struct {
	Header map[string]interface{}
	//Len    uint32 // len(dip) + len(dport) + + len(t) + len(pip) + len(pport)...
	T       []byte `json:"t"`     // 1byte, 0: req; 1:resp
	Dip     []byte `json:"dip"`   // 4byte, 0-255 for every node
	Dport   []byte `json:"dport"` // 2byte, 0-65535
	User    []byte `json:"user"`
	Pass    []byte `json:"pass"`
	HostKey []byte `json:"host_key"`
	Ppip    []byte `json:"ppip"`   // proxy server public ip
	Pip     []byte `json:"pip"`    // proxy server private ip
	Pport   []byte `json:"pport"`  // proxy server port
	Prport  []byte `json:"prport"` // proxy listener random port
	Err     []byte `json:"err"`    // empty: everything is ok; not empty: error occurred
}

func NewMessage() *SProtocol {
	return &SProtocol{}
}

func (proto SProtocol) Valid() (bool, error) {

	if len(proto.Dip) != 4 {
		return false, fmt.Errorf("server ip error")
	}

	if len(proto.Dport) > 2 || len(proto.Dport) <= 0 {
		return false, fmt.Errorf("server port is big than 65535 or zero")
	}

	if len(proto.Ppip) != 4 {
		return false, fmt.Errorf("proxy public ip error")
	}

	if len(proto.Pip) != 4 {
		return false, fmt.Errorf("proxy private ip error")
	}

	if len(proto.Pport) > 2 || len(proto.Pport) <= 0 {
		return false, fmt.Errorf("proxy port is big than 65535 or zero")
	}

	if len(proto.T) != 1 {
		return false, fmt.Errorf("type length error")
	}

	if len(proto.User) <= 0 {
		return false, fmt.Errorf("username cann't be empty")
	}

	//if proto.Len != uint32(len(proto.t)+len(proto.dip)+len(proto.dport)+len(proto.ppip)+len(proto.pip)+len(proto.pport)+
	//	len(proto.user)+len(proto.pass)+len(proto.prport)) {
	//	return false, fmt.Errorf("message length error")
	//}

	return true, nil
}

func (proto SProtocol) IsReq() bool {
	return proto.T[0] == uint8(0)
}

func (proto SProtocol) IsResp() bool {
	return proto.T[0] == uint8(1)
}

// Getters
func (proto *SProtocol) GetUser() string {
	return string(proto.User)
}

func (proto *SProtocol) GetPass() string {
	return string(proto.Pass)
}

func (proto *SProtocol) GetKeySig() ssh.Signer {
	var hostKey utils.HostKey
	hostKey.Value = string(proto.HostKey)
	sig, err := hostKey.Load()
	if err != nil {
		common.Log.Errorf("couldn't load host key from value: %s", err.Error())
		return nil
	}
	return sig
}

func (proto SProtocol) GetDip() string {
	return fmt.Sprintf("%d.%d.%d.%d", proto.Dip[0], proto.Dip[1], proto.Dip[2], proto.Dip[3])
}

func (proto SProtocol) GetDPort() uint16 {
	if len(proto.Dport) == 1 {
		return uint16(proto.Dport[0])
	} else {
		return (uint16(proto.Dport[0]) << 8) | (uint16(proto.Dport[1]) & 0xFF)
	}
}

func (proto SProtocol) GetPPip() string {
	return fmt.Sprintf("%d.%d.%d.%d", proto.Ppip[0], proto.Ppip[1], proto.Ppip[2], proto.Ppip[3])
}

func (proto SProtocol) GetPip() string {
	return fmt.Sprintf("%d.%d.%d.%d", proto.Pip[0], proto.Pip[1], proto.Pip[2], proto.Pip[3])
}

func (proto SProtocol) GetPPort() uint16 {
	if len(proto.Pport) == 1 {
		return uint16(proto.Pport[0])
	} else {
		return (uint16(proto.Pport[0]) << 8) | (uint16(proto.Pport[1]) & 0xFF)
	}
}

func (proto SProtocol) GetPRPort() uint16 {
	if len(proto.Prport) == 1 {
		return uint16(proto.Prport[0])
	} else {
		return (uint16(proto.Prport[0]) << 8) | (uint16(proto.Prport[1]) & 0xFF)
	}
}

func (proto SProtocol) GetErr() error {
	if len(proto.Err) == 0 {
		return nil
	} else {
		return fmt.Errorf(string(proto.Err))
	}
}

// Setters
func (proto *SProtocol) SetT(t uint8) {
	proto.T = []byte{t}
}

func (proto *SProtocol) SetDip(ip string) error {
	var dip []byte
	for _, n := range strings.Split(ip, ".") {
		_n, err := strconv.Atoi(n)
		if err != nil {
			return err
		}
		dip = append(dip, uint8(_n))
	}
	if len(dip) != 4 {
		return fmt.Errorf("ip lenght error")
	}
	proto.Dip = dip
	return nil
}

func (proto *SProtocol) SetDPort(port uint16) {
	proto.Dport = []byte{uint8(port >> 8), uint8(port & 0xFF)}
}

func (proto *SProtocol) SetPPip(ip string) error {
	var ppip []byte
	for _, n := range strings.Split(ip, ".") {
		_n, err := strconv.Atoi(n)
		if err != nil {
			return err
		}
		ppip = append(ppip, uint8(_n))
	}
	if len(ppip) != 4 {
		return fmt.Errorf("proxy public ip length error")
	}
	proto.Ppip = ppip
	return nil
}

func (proto *SProtocol) SetPip(ip string) error {
	var pip []byte
	for _, n := range strings.Split(ip, ".") {
		_n, err := strconv.Atoi(n)
		if err != nil {
			return err
		}
		pip = append(pip, uint8(_n))
	}
	if len(pip) != 4 {
		return fmt.Errorf("proxy private ip length error")
	}
	proto.Pip = pip
	return nil
}

func (proto *SProtocol) SetPPort(port uint16) {
	proto.Pport = []byte{uint8(port >> 8), uint8(port & 0xFF)}
}

func (proto *SProtocol) SetPRPort(port uint16) {
	proto.Prport = []byte{uint8(port >> 8), uint8(port & 0xFF)}
}

func (proto *SProtocol) SetUser(user string) {
	proto.User = []byte(user)
}

func (proto *SProtocol) SetPass(pass string) {
	proto.Pass = []byte(pass)
}

func (proto *SProtocol) SetKeySig(ks utils.HostKey) {
	if len(ks.Value) != 0 {
		proto.HostKey = []byte(ks.Value)
		return
	}
	ksf, err := os.Open(ks.Path)
	if err != nil {
		common.Log.Errorf("failed to load host key from file: %s", ks.Path)
		return
	}
	var bf = make([]byte, 2048)
	n, err := ksf.Read(bf)
	if err != nil {
		common.Log.Errorf("failed to load host key from file: %s", ks.Path)
		return
	}
	proto.HostKey = bf[:n]
}

func (proto *SProtocol) SetErr(err string) {
	proto.Err = []byte(err)
}
