package router

import (
	"github.com/gin-gonic/gin"
	"zeus/controllers"
)

func userRouter(r *gin.Engine) {
	userR := r.Group("/api/v1/users")
	{
		userR.GET("", controllers.FetchUserList)
		userR.PUT("/:username/valid", controllers.ValidUser)
	}
}
