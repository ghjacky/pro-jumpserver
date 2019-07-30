package router

import "github.com/gin-gonic/gin"

var R *gin.Engine

func init() {
	R = gin.Default()
}

// register 注册子路由
func register() {

}
