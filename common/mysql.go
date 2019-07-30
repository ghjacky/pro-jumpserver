package common

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"strings"
)

type mysqlConfig struct {
	host       string
	port       int
	user       string
	password   string
	database   string
	connParams []string
}

var Mysql *gorm.DB

func initMysql() {
	var err error
	Log.Infoln("Connecting Mysql ......")
	Mysql, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		Config.mysqlConfig.user, Config.mysqlConfig.password, Config.mysqlConfig.host, Config.mysqlConfig.port,
		Config.mysqlConfig.database, strings.Join(Config.mysqlConfig.connParams, "&")))
	if err != nil {
		Log.Fatalf("Couldn't to mysql at %s:%d", Config.mysqlConfig.host, Config.mysqlConfig.port)
	} else {
		Log.Infof("Connected to mysql at %s:%d successfully", Config.mysqlConfig.host, Config.mysqlConfig.port)
		Log.Infoln("Config mysql to LogMode")
		Mysql = Mysql.LogMode(true)
	}
}
