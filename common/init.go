package common

// Init 初始化common下所有相应组件
func Init() {
	// 配置初始化
	Config.initConfig()
	// 初始化Log
	initLog()
	// 初始化mysql连接
	initMysql()
	// 初始化redis连接
	initRedis()
	// 初始化ldap连接
	InitLdap()
}
