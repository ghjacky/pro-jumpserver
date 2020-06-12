package common

import (
	"fmt"
	"github.com/go-ldap/ldap/v3"
)

type ldapConfig struct {
	server      string
	port        int
	dn          string
	searchScope string
	bindUser    string
	password    string
}

var LdapConn *ldap.Conn

func initLdap() {
	var err error
	LdapConn, err = ldap.Dial("tcp", fmt.Sprintf("%s:%d", Config.ldapConfig.server, Config.ldapConfig.port))
	if err != nil {
		Log.Errorf("Connect to ldap: %s error: %s", Config.ldapConfig.server, err.Error())
		return
	} else {
		if err = LdapConn.Bind(Config.ldapConfig.bindUser, Config.ldapConfig.password); err != nil {
			Log.Errorf("Couldn't bind user: %s to ldap: %s, error: %s", Config.ldapConfig.bindUser, Config.ldapConfig.server, err.Error())
			return
		}
	}
	Log.Infof("Connected to ldap: %s:%d", Config.ldapConfig.server, Config.ldapConfig.port)
	return
}
