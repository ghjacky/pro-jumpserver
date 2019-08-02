package router

import "github.com/gin-gonic/gin"

func permissionRouter(r *gin.Engine) {
	permR := r.Group("/api/v1/")
	{
		permR.GET("", getUserPerm)
		permR.POST("", addUserPerm)
		permR.PUT("", setUserPerm)
	}
}
