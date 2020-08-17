package controllers

import (
	"github.com/gin-gonic/gin"
	"zeus/modules/jumpserver"
)

func GetPubKey(ctx *gin.Context) {
	pubKey := jumpserver.GetPubKey()
	if len(pubKey) == 0 {
		ctx.JSON(200, newHttpResp(100002, "public key 获取出错", nil))
		return
	}
	ctx.JSON(200, newHttpResp(100000, "public 获取成功", pubKey))
	return
}
