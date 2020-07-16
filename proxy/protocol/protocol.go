package protocol

import (
	"fmt"
	"strconv"
	"strings"
)

type SProtocol struct {
	Header map[string]interface{}
	Len    uint32 // len(dip) + len(dport) + + len(t) + len(pip) + len(pport)
	t      []byte // 1byte, 0: req; 1:resp
	dip    []byte // 4byte, 0-255 for every node
	dport  []byte // 2byte, 0-65535
	user   []byte
	pass   []byte
	ppip   []byte // proxy server public ip
	pip    []byte // proxy server private ip
	pport  []byte // proxy server port
	prport []byte // proxy listener random port
	err    []byte // empty: everything is ok; not empty: error occurred
}

func NewMessage() *SProtocol {
	return &SProtocol{}
}

func (proto SProtocol) Valid() (bool, error) {

	if len(proto.dip) != 4 {
		return false, fmt.Errorf("server ip error")
	}

	if len(proto.dport) > 2 || len(proto.dport) <= 0 {
		return false, fmt.Errorf("server port is big than 65535 or zero")
	}

	if len(proto.ppip) != 4 {
		return false, fmt.Errorf("proxy public ip error")
	}

	if len(proto.pip) != 4 {
		return false, fmt.Errorf("proxy private ip error")
	}

	if len(proto.pport) > 2 || len(proto.pport) <= 0 {
		return false, fmt.Errorf("proxy port is big than 65535 or zero")
	}

	if len(proto.t) != 1 {
		return false, fmt.Errorf("type length error")
	}

	if len(proto.user) <= 0 {
		return false, fmt.Errorf("username cann't be empty")
	}

	if proto.Len != uint32(len(proto.t)+len(proto.dip)+len(proto.dport)+len(proto.ppip)+len(proto.pip)+len(proto.pport)+
		len(proto.user)+len(proto.pass)+len(proto.prport)) {
		return false, fmt.Errorf("message length error")
	}

	return true, nil
}

func (proto SProtocol) IsReq() bool {
	return proto.t[0] == uint8(0)
}

func (proto SProtocol) IsResp() bool {
	return proto.t[0] == uint8(1)
}

// Getters
func (proto *SProtocol) GetUser() string {
	return string(proto.user)
}

func (proto *SProtocol) GetPass() string {
	return string(proto.pass)
}

func (proto SProtocol) GetDip() string {
	return fmt.Sprintf("%d.%d.%d.%d", proto.dip[0], proto.dip[1], proto.dip[2], proto.dip[3])
}

func (proto SProtocol) GetDPort() uint16 {
	if len(proto.dport) == 1 {
		return uint16(proto.dport[0])
	} else {
		return (uint16(proto.dport[0]) << 8) | (uint16(proto.dport[1]) & 0xFF)
	}
}

func (proto SProtocol) GetPPip() string {
	return fmt.Sprintf("%d.%d.%d.%d", proto.ppip[0], proto.ppip[1], proto.ppip[2], proto.ppip[3])
}

func (proto SProtocol) GetPip() string {
	return fmt.Sprintf("%d.%d.%d.%d", proto.pip[0], proto.pip[1], proto.pip[2], proto.pip[3])
}

func (proto SProtocol) GetPPort() uint16 {
	if len(proto.pport) == 1 {
		return uint16(proto.pport[0])
	} else {
		return (uint16(proto.pport[0]) << 8) | (uint16(proto.pport[1]) & 0xFF)
	}
}

func (proto SProtocol) GetPRPort() uint16 {
	if len(proto.prport) == 1 {
		return uint16(proto.prport[0])
	} else {
		return (uint16(proto.prport[0]) << 8) | (uint16(proto.prport[1]) & 0xFF)
	}
}

func (proto SProtocol) GetErr() error {
	if len(proto.err) == 0 {
		return nil
	} else {
		return fmt.Errorf(string(proto.err))
	}
}

// Setters
func (proto *SProtocol) SetT(t uint8) {
	proto.t = []byte{t}
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
	proto.dip = dip
	return nil
}

func (proto *SProtocol) SetDPort(port uint16) {
	proto.dport = []byte{uint8(port >> 8), uint8(port & 0xFF)}
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
	proto.ppip = ppip
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
	proto.pip = pip
	return nil
}

func (proto *SProtocol) SetPPort(port uint16) {
	proto.pport = []byte{uint8(port >> 8), uint8(port & 0xFF)}
}

func (proto *SProtocol) SetPRPort(port uint16) {
	proto.prport = []byte{uint8(port >> 8), uint8(port & 0xFF)}
}

func (proto *SProtocol) SetErr(err string) {
	proto.err = []byte(err)
}
