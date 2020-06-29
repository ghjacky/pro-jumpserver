package main

import (
	"flag"
	"time"
	"zeus/common"
	"zeus/modules/jumpserver"
	"zeus/modules/webserver/users"
	"zeus/router"
)

func bgJobs() {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			if err := users.FetchUserFromLDAP(); err != nil {
				common.Log.Errorf("从ldap同步用户数据出错：%s", err.Error())
			}
		}
	}()
}

func main() {
	// 定义命令行参数'--config'， 默认值为"./configs/config.toml"
	common.Configfile = flag.String("config", "./configs/config.toml", "Specify config file for server")
	flag.Parse()
	initAll()
	go func() {
		if err := router.R.Run(common.Config.WebServerAddr); err != nil {
			common.Log.Errorf("无法启动web服务:%s", err.Error())
		}
	}()
	defer common.Exit()
	bgJobs()
	//jumpserver.GenGACqr()
	// 运行jumpserver
	jumpserver.InitJumpServer()
	defer jumpserver.JPS.Stop()
	jumpserver.JPS.Run()
}
