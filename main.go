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
	initAll()
	defer common.Exit()
	//jumpserver.GenGACqr()
	// 运行jumpserver
	jumpserver.InitJumpServer()
	defer jumpserver.JPS.Stop()
	jumpserver.JPS.Run()
}
