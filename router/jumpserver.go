package router

import (
	"github.com/gin-gonic/gin"
	"zeus/controllers"
)

func jsRouter(r *gin.Engine) {
	jsR := r.Group("/api/v1/keys")
	{
		jsR.GET("/public", controllers.GetPubKey)
	}
}
