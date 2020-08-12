package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"
	"zeus/common"
	"zeus/models"
	"zeus/modules/jumpserver"
	"zeus/modules/webserver/user"
	"zeus/modules/wsserver"
	"zeus/router"
)

func bgJobs() {
	go catchOsSignal()
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			if err := user.FetchUserFromLDAP(); err != nil {
				common.Log.Errorf("从ldap同步用户数据出错：%s", err.Error())
			}
		}
	}()
}

func catchOsSignal() {
	common.Mysql.Exec("update users set active = ?", models.UserActiveNo)
	common.Log.Debugln("开始监听系统信号")
	sig := make(chan os.Signal, 0)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGSTOP)
	<-sig
	common.Mysql.Exec("update users set active = ?", models.UserActiveNo)
	common.Log.Debugln("监听到系统推出信号")
	os.Exit(0)
}

func main() {
	// 定义命令行参数'--config'， 默认值为"./configs/config.toml"
	common.Configfile = flag.String("config", "./configs/config.toml", "Specify config file for server")
	flag.Parse()
	initAll()
	// 运行web server
	go func() {
		if err := router.R.Run(common.Config.WebServerAddr); err != nil {
			common.Log.Errorf("无法启动web服务:%s", err.Error())
		}
	}()
	go monitorOsSignal()
	// 运行websocket server
	go wsserver.Run()
	defer common.Exit()
	bgJobs()
	//jumpserver.GenGACqr()
	// 运行jumpserver
	jumpserver.InitJumpServer()
	defer jumpserver.JPS.Stop()
	jumpserver.JPS.Run()
}
