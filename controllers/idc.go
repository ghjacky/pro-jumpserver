package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"zeus/models"
)

func FetchAllIDCS(ctx *gin.Context) {
	var idcs = new(models.SIDCS)
	if err := idcs.FetchAll(); err != nil {
		ctx.JSON(200, newHttpResp(100002, fmt.Sprintf("数据库错误：%s", err.Error()), nil))
		return
	}
	ctx.JSON(200, newHttpResp(100000, "成功获取IDC列表", idcs))
	return
}

func AddIDC(ctx *gin.Context) {
	var idc = new(models.SIDC)
	if err := ctx.BindJSON(idc); err != nil {
		ctx.JSON(200, newHttpResp(100001, fmt.Sprintf("参数错误：%s", err.Error()), nil))
		return
	}
	if err := idc.UpdateOrAdd(); err != nil {
		ctx.JSON(200, newHttpResp(100002, fmt.Sprintf("数据库错误：%s", err.Error()), nil))
		return
	}
	ctx.JSON(200, newHttpResp(100000, "idc添加成功", idc))
	return
}

func DeleteIDC(ctx *gin.Context) {
	var idc = &models.SIDC{Name: ctx.Param("name")}
	if err := idc.Delete(); err != nil {
		ctx.JSON(200, newHttpResp(100001, fmt.Sprintf("数据库错误：%s", err.Error()), nil))
		return
	}
	ctx.JSON(200, newHttpResp(100000, "idc删除成功", idc))
	return
}

func SetProxy(ctx *gin.Context) {
	var idc = new(models.SIDC)
	if err := ctx.BindJSON(idc); err != nil {
		ctx.JSON(200, newHttpResp(100001, fmt.Sprintf("参数错误：%s", err.Error()), nil))
		return
	}
	if err := idc.UpdateOrAdd(); err != nil {
		ctx.JSON(200, newHttpResp(100002, fmt.Sprintf("数据库错误：%s", err.Error()), nil))
		return
	}
	ctx.JSON(200, newHttpResp(100000, "代理设置成功", idc))
	return
}
