package router

import (
	"github.com/gin-gonic/gin"
	"zeus/controllers"
)

func permissionRouter(r *gin.Engine) {
	permR := r.Group("/api/v1/perms")
	{
		permR.GET("/:username", controllers.GetUserPerm)
		permR.POST("", controllers.AddUserPerm)
		permR.PUT("", controllers.SetUserPerm)
	}
}
