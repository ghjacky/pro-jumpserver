package utils

import (
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"path"
)

type HostKey struct {
	Value string
	Path  string
}

func (hk *HostKey) loadHostKeyFromFile(keyPath string) (signer ssh.Signer, err error) {
	_, err = os.Stat(keyPath)
	if err != nil {
		return
	}
	buf, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return
	}
	return hk.loadHostKeyFromString(string(buf))
}

func (hk *HostKey) loadHostKeyFromString(value string) (signer ssh.Signer, err error) {
	signer, err = ssh.ParsePrivateKey([]byte(value))
	return
}

func (hk *HostKey) Gen() (signer ssh.Signer, err error) {
	key, err := GeneratePrivateKey(2048)
	if err != nil {
		return
	}
	keyBytes := EncodePrivateKeyToPEM(key)
	keyDir := path.Dir(hk.Path)
	if !FileExists(keyDir) {
		err := os.MkdirAll(keyDir, os.ModePerm)
		if err != nil {
			return signer, err
		}
	}
	err = WriteKeyToFile(keyBytes, hk.Path)
	if err != nil {
		return
	}
	return ssh.NewSignerFromKey(key)
}

func (hk *HostKey) Load() (signer ssh.Signer, err error) {
	if hk.Value != "" {
		signer, err = hk.loadHostKeyFromString(hk.Value)
		if err == nil {
			return
		}
	}
	if hk.Path != "" {
		signer, err = hk.loadHostKeyFromFile(hk.Path)
		if err == nil {
			return
		}
	}
	signer, err = hk.Gen()
	return
}
