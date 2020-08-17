package router

import (
	"github.com/gin-gonic/gin"
	"zeus/middleware"
)

var R *gin.Engine

func Init() {
	R = gin.Default()
	R.Use(middleware.CheckToken())
	Register()
}

// register 注册子路由
func Register() {
	permissionRouter(R)
	userRouter(R)
	eventRouter(R)
	idcRouter(R)
	jsRouter(R)
}
