package utils

import (
	"fmt"
	"net"
	"zeus/common"
)

const (
	MAXSENDSIZE = 65535
	BUFFERSIZE  = 256
)

func SendTo(c net.Conn, data []byte) error {
	sizeHolder := make([]byte, 2)
	dataLen := len(data)
	if dataLen > MAXSENDSIZE {
		return fmt.Errorf("too big data")
	}
	sizeHolder = []byte{uint8(dataLen >> 8), uint8(dataLen & 0xff)}
	if n, err := c.Write(sizeHolder); err != nil {
		return err
	} else if n != 2 {
		return fmt.Errorf("wrong data size")
	}
	if n, err := c.Write(data); err != nil {
		return err
	} else if n != len(data) {
		return fmt.Errorf("wrong data send size")
	}
	return nil
}

func RecvFrom(c net.Conn, recvChan *TimeoutChan) {
	buf, tbf, sbf := genBuf()
	for {
		size, err := readSize(c, sbf)
		if err != nil {
			common.Log.Warnf("read data size from socket error: %s", err.Error())
			break
		}
		if size == 0 {
			continue
		}
		for {
			n, err := c.Read(tbf)
			if err != nil {
				common.Log.Warnf("read data from socket error: %s", err.Error())
				break
			}
			buf = append(buf, tbf[:n]...)
			size -= uint16(n)
			if size == 0 {
				if timeout := recvChan.WriteWithTimeout(buf); timeout {
					break
				}
				buf, tbf, sbf = genBuf()
				break
			}
		}
	}
}

func genBuf() (buf, tbf, sbf []byte) {
	buf = make([]byte, 0)
	tbf = make([]byte, BUFFERSIZE)
	sbf = make([]byte, 2)
	return
}

func readSize(c net.Conn, sbf []byte) (uint16, error) {
	if n, err := c.Read(sbf); err != nil {
		return 0, err
	} else if n != 2 {
		return 0, nil
	}
	return uint16(sbf[0])<<8 | uint16(sbf[1]), nil
}
