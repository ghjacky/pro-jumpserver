package router

import "github.com/gin-gonic/gin"

var R *gin.Engine

func Init() {
	R = gin.Default()
	Register()
}

// register 注册子路由
func Register() {
	permissionRouter(R)
}
