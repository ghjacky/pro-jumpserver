package main

import (
	"os"
	"os/signal"
	"syscall"
	"zeus/common"
	"zeus/models"
	"zeus/router"
)

func initAll() {
	// 初始化程序配置，mysql、redis连接以及日志配置等
	common.Init()
	common.Mysql.AutoMigrate(
		&models.Event{},
		&models.User{},
		&models.Server{},
		&models.Asset{},
		&models.SIDC{},
		&models.SProxy{})
	// -- temp test code
	var p = models.SProxy{
		IDC:  "北京",
		PIP:  []byte{192, 168, 32, 75},
		PPIP: []byte{106, 12, 80, 193},
		//PIP:   []byte{192, 168, 72, 138},
		//PPIP:  []byte{127, 0, 0, 1},
		PPORT: 2000,
	}
	p.Add()
	// --
	router.Init()
}

func monitorOsSignal() {
	sc := make(chan os.Signal, 0)
	signal.Notify(sc, syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGHUP, syscall.SIGQUIT)
	for {
		_ = <-sc
		os.Exit(0)
	}
}
