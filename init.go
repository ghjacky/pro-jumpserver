package main

import (
	"os"
	"os/signal"
	"syscall"
	"zeus/common"
	"zeus/models"
	"zeus/modules/jumpserver"
	"zeus/router"
)

func initAll() {
	// 初始化程序配置，mysql、redis连接以及日志配置等
	common.Init()
	common.Mysql.AutoMigrate(&models.Event{}, &models.User{}, &models.Server{}, &models.Asset{}, &models.SIDC{}, &models.SProxy{})
	router.Init()
}

func monitorOsSignal() {
	sc := make(chan os.Signal, 0)
	signal.Notify(sc, syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGHUP, syscall.SIGQUIT)
	for {
		_ = <-sc
		// 系统退出前清除所有连接和后台任务
		clearAll()
		os.Exit(0)
	}
}

func clearAll() {
	jumpserver.ExitSessionBgTask(100)
}
