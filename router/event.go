package router

import (
	"github.com/gin-gonic/gin"
	"zeus/controllers"
)

func eventRouter(r *gin.Engine) {
	eventR := r.Group("/api/v1/events")
	{
		eventR.GET("", controllers.FetchEvents)
	}
}
