package router

import (
	"github.com/gin-gonic/gin"
	"zeus/controllers"
)

func permissionRouter(r *gin.Engine) {
	permR := r.Group("/api/v1/perms")
	{
		permR.GET("", controllers.FetchPermissions)
		permR.GET("/:username", controllers.GetUserPermissions)
		permR.POST("/:username", controllers.AddUserPermissions)
		permR.PUT("", controllers.UpdateUserPermissions)
		permR.DELETE("/:id", controllers.DeletePermission)
	}
}
