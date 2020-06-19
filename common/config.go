package common

import (
	"github.com/spf13/viper"
	"os"
	"strings"
)

// 定义部分默认配置
const (
	DefaultWebServerAddr  = "0.0.0.0:8080"
	DefaultJumpServerAddr = "0.0.0.0:2222"
	DefaultWorkDir        = "/tmp/var/lib/zeus"
	DefaultDataDir        = "/tmp/var/lib/zeus/data"
	DefaultPlayDir        = "/tmp/var/lib/zeus/play"
	DefaultLogFile        = "/tmp/var/lib/zeus/server.log"
	DefaultMysqlHost      = "127.0.0.1"
	DefaultMysqlPort      = 3306
	DefaultMysqlUser      = "root"
	DefaultMysqlPassword  = "roothjack"
	DefaultMysqlDatabase  = "zeus"
	DefaultRedisHost      = "127.0.0.1"
	DefaultRedisPort      = 6379
	DefaultRedisPassword  = ""
	DefaultRedisDb        = 0
	DefaultFileMode       = 0644
	DefaultDirMode        = 0755
)

type mainConfig struct {
	WebServerAddr  string
	JumpServerAddr string
	WorkDir        string
	DataDir        string
	PlayDir        string
	LogFile        *os.File
	IDCs           []string
}
type config struct {
	mainConfig
	mysqlConfig
	redisConfig
	ldapConfig
	HeraConfig
}

var Config = &config{}
var Configfile *string

// configuration 读取配置文件，初始化配置
func (c *config) initConfig() {
	var err error
	viper.SetConfigFile(*Configfile)
	viper.SetConfigType("toml")
	logfile := ""
	if err := viper.ReadInConfig(); err != nil {
		Log.Warnf("Couldn't read config file %s", *Configfile)
		Log.Warnf("Using default configuration!")
	} else {
		c.WebServerAddr = viper.GetString("main.webServerAddr")
		c.JumpServerAddr = viper.GetString("main.JumpServerAddr")
		c.WorkDir = viper.GetString("main.workDir")
		c.DataDir = viper.GetString("main.dataDir")
		c.PlayDir = viper.GetString("main.playDir")
		logfile = viper.GetString("main.logfile")
		c.IDCs = strings.Split(viper.GetString("main.idcs"), ",")
		c.mysqlConfig.user = viper.GetString("mysql.user")
		c.mysqlConfig.host = viper.GetString("mysql.host")
		c.mysqlConfig.port = viper.GetInt("mysql.port")
		c.mysqlConfig.database = viper.GetString("mysql.database")
		c.mysqlConfig.password = viper.GetString("mysql.password")
		c.redisConfig.host = viper.GetString("redis.host")
		c.redisConfig.port = viper.GetInt("redis.port")
		c.redisConfig.password = viper.GetString("redis.password")
		c.redisConfig.db = viper.GetInt("redis.db")
		c.ldapConfig.server = viper.GetString("ldap.server")
		c.ldapConfig.port = viper.GetInt("ldap.port")
		c.ldapConfig.dn = viper.GetString("ldap.dn")
		c.ldapConfig.searchScope = viper.GetString("ldap.search_scope")
		c.ldapConfig.bindUser = viper.GetString("ldap.bind_user")
		c.ldapConfig.password = viper.GetString("ldap.password")
		c.HeraConfig.Name = viper.GetString("hera.name")
		c.HeraConfig.Addr = strings.Trim(viper.GetString("hera.addr"), "/")
		c.HeraConfig.ApiPrefix = strings.Trim(viper.GetString("hera.api_prefix"), "/")
	}
	if len(logfile) == 0 {
		logfile = DefaultLogFile
	}
	c.checkAndSetDefault()
	c.createIfNotExist()
	c.LogFile, err = os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, DefaultFileMode)
	if err != nil {
		Log.Fatalf("Couldn't open log file: %s", logfile)
	}
	c.mysqlConfig.connParams = []string{"charset=utf8", "parseTime=True"}
	Log.Info("Init configuration successfully !")
}

// checkAndSetDefault 检测配置项是否为0值，如果为0值则设置默认值
func (c *config) checkAndSetDefault() {
	if len(c.mainConfig.JumpServerAddr) == 0 {
		c.mainConfig.JumpServerAddr = DefaultJumpServerAddr
	}
	if len(c.mainConfig.WebServerAddr) == 0 {
		c.mainConfig.WebServerAddr = DefaultWebServerAddr
	}
	if len(c.mainConfig.WorkDir) == 0 {
		c.mainConfig.WorkDir = DefaultWorkDir
	}
	if len(c.mainConfig.DataDir) == 0 {
		c.mainConfig.DataDir = DefaultDataDir
	}
	if len(c.mainConfig.PlayDir) == 0 {
		c.mainConfig.PlayDir = DefaultPlayDir
	}
	if len(c.mysqlConfig.host) == 0 {
		c.mysqlConfig.host = DefaultMysqlHost
	}
	if c.mysqlConfig.port == 0 {
		c.mysqlConfig.port = DefaultMysqlPort
	}
	if len(c.mysqlConfig.user) == 0 {
		c.mysqlConfig.user = DefaultMysqlUser
	}
	if len(c.mysqlConfig.password) == 0 {
		c.mysqlConfig.password = DefaultMysqlPassword
	}
	if len(c.mysqlConfig.database) == 0 {
		c.mysqlConfig.database = DefaultMysqlDatabase
	}
	if len(c.redisConfig.host) == 0 {
		c.redisConfig.host = DefaultRedisHost
	}
	if c.redisConfig.port == 0 {
		c.redisConfig.port = DefaultRedisPort
	}
	if len(c.redisConfig.password) == 0 {
		c.redisConfig.password = DefaultRedisPassword
	}
	if c.redisConfig.db == 0 {
		c.redisConfig.db = DefaultRedisDb
	}
}

// createIfNotExist 检测配置中等文件或目录是否存在，不存在则创建，创建失败则退出
func (c *config) createIfNotExist() {
	if err := os.MkdirAll(c.PlayDir, DefaultDirMode); err != nil {
		Log.Fatalf("couldn't create play dir: %s", c.PlayDir)
	}
	if err := os.MkdirAll(c.DataDir, DefaultDirMode); err != nil {
		Log.Fatalf("couldn't create data dir: %s", c.DataDir)
	}
	if err := os.MkdirAll(c.WorkDir, DefaultDirMode); err != nil {
		Log.Fatalf("couldn't create work dir: %s", c.WorkDir)
	}
}
