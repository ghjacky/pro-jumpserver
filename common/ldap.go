package common

import (
	"fmt"
	"github.com/go-ldap/ldap/v3"
)

type LdapConfig struct {
	Server      string
	Port        int
	Dn          string
	SearchScope string
	BindUser    string
	Password    string
}

var LdapConn *ldap.Conn

func initLdap() {
	var err error
	LdapConn, err = ldap.Dial("tcp", fmt.Sprintf("%s:%d", Config.LdapConfig.Server, Config.LdapConfig.Port))
	if err != nil {
		Log.Errorf("Connect to ldap: %s error: %s", Config.LdapConfig.Server, err.Error())
		return
	} else {
		if err = LdapConn.Bind(Config.LdapConfig.BindUser, Config.LdapConfig.Password); err != nil {
			Log.Errorf("Couldn't bind user: %s to ldap: %s, error: %s", Config.LdapConfig.BindUser, Config.LdapConfig.Server, err.Error())
			return
		}
	}
	Log.Infof("Connected to ldap: %s:%d", Config.LdapConfig.Server, Config.LdapConfig.Port)
	return
}
