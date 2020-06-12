package common

func Exit() {
	// 关闭日志文件
	if err := Config.LogFile.Close(); err != nil {
		Log.Errorf("Couldn't closing log file: %s", Config.LogFile.Name())
	}
	// 关闭mysql连接
	if err := Mysql.Close(); err != nil {
		Log.Errorf("Couldn't closing mysql connection to %s:%d", Config.mysqlConfig.host, Config.mysqlConfig.port)
	}
	// 关闭redis连接
	if err := Redis.Close(); err != nil {
		Log.Errorf("Couldn't closing redis connection to %s:%d", Config.redisConfig.host, Config.redisConfig.port)
	}
	// 关闭ldap连接
	LdapConn.Close()
}
