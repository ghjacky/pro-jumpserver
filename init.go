package main

import (
	"zeus/common"
	"zeus/models"
	"zeus/router"
)

func initAll() {
	// 初始化程序配置，mysql、redis连接以及日志配置等
	common.Init()
	common.Mysql.AutoMigrate(&models.Event{}, &models.User{}, &models.Server{}, &models.Asset{})
	router.Init()
}
