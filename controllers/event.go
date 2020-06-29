package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"zeus/models"
	"zeus/modules/webserver/events"
)

func FetchEvents(ctx *gin.Context) {
	query := models.Query{}
	if err := ctx.BindQuery(&query); err != nil {
		ctx.JSON(200, newHttpResp(100001, fmt.Sprintf("参数错误: %#v", query), nil))
		return
	}
	total, evs, err := events.FetchEventList(query)
	if err != nil {
		ctx.JSON(200, newHttpResp(100302, fmt.Sprintf("获取事件列表出错：%s", err.Error()), nil))
		return
	}
	ctx.JSON(200, newHttpResp(100000, "成功获取事件列表", map[string]interface{}{"total": total, "events": evs}))
	return
}

func PlayEvent(ctx *gin.Context) {

}
