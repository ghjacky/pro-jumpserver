package main

import (
	"flag"
	"zeus/common"
	"zeus/modules/jumpserver"
)

func main() {
	// 定义命令行参数'--config'， 默认值为"./configs/config.toml"
	common.Configfile = flag.String("config", "./configs/config.toml", "Specify config file for server")
	flag.Parse()
	// 初始化程序配置，mysql、redis连接以及日志配置等
	common.Init()
	defer common.Exit()
	//jumpserver.GenGACqr()
	// 运行jumpserver
	defer jumpserver.JPS.Stop()
	jumpserver.JPS.Run()
}
