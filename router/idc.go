package router

import (
	"github.com/gin-gonic/gin"
	"zeus/controllers"
)

func idcRouter(r *gin.Engine) {
	idcR := r.Group("/api/v1/idcs")
	{
		idcR.GET("", controllers.FetchAllIDCS)
		idcR.POST("", controllers.AddIDC)
		idcR.DELETE("/:name", controllers.DeleteIDC)
		idcR.PUT("", controllers.SetProxy)
	}
	proxyR := r.Group("/api/v1/proxies")
	{
		proxyR.GET("", controllers.FetchAllProxies)
		proxyR.POST("", controllers.AddProxy)
		proxyR.PUT("", controllers.UpdateProxy)
		proxyR.DELETE("/:id", controllers.DeleteProxy)
	}
}
