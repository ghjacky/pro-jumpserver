package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"zeus/models"
	"zeus/utils"
)

func AddProxy(ctx *gin.Context) {
	var proxy = new(models.SProxy)
	if err := ctx.BindJSON(proxy); err != nil {
		ctx.JSON(200, newHttpResp(100001, fmt.Sprintf("参数错误：%s", err.Error()), nil))
		return
	}
	if err := proxy.Add(); err != nil {
		ctx.JSON(200, newHttpResp(100002, fmt.Sprintf("数据库错误: %s", err.Error()), nil))
		return
	}
	ctx.JSON(200, newHttpResp(100000, "成功添加代理", proxy))
	return
}

func UpdateProxy(ctx *gin.Context) {
	var proxy = new(models.SProxy)
	if err := ctx.BindJSON(proxy); err != nil {
		ctx.JSON(200, newHttpResp(100001, fmt.Sprintf("参数错误：%s", err.Error()), nil))
		return
	}
	if err := proxy.Update(); err != nil {
		ctx.JSON(200, newHttpResp(100002, fmt.Sprintf("数据库错误: %s", err.Error()), nil))
		return
	}
	ctx.JSON(200, newHttpResp(100000, "成功更新代理", proxy))
	return
}

func DeleteProxy(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(200, newHttpResp(100001, fmt.Sprintf("参数错误：%s", err.Error()), nil))
		return
	}
	var proxy = new(models.SProxy)
	proxy.ID = uint(id)
	if err := proxy.Delete(); err != nil {
		ctx.JSON(200, newHttpResp(100002, fmt.Sprintf("数据库错误：%s", err.Error()), nil))
		return
	}
	ctx.JSON(200, newHttpResp(100000, "成功删除代理", proxy))
	return
}

func FetchAllProxies(ctx *gin.Context) {
	var proxies = new(models.SProxies)
	if err := proxies.FetchAll(); err != nil {
		ctx.JSON(200, newHttpResp(100001, fmt.Sprintf("数据库错误：%s", err.Error()), nil))
		return
	}
	var pmaps []interface{}
	for _, proxy := range *proxies {
		p, err := utils.StructToMap(proxy)
		if err != nil {
			ctx.JSON(200, newHttpResp(100000, "成功获取代理列表", proxies))
			return
		}
		var pmap = p.(map[string]interface{})
		pmap["ppip"] = utils.ByteArrToStringForIP(proxy.PPIP)
		pmap["pip"] = utils.ByteArrToStringForIP(proxy.PIP)
		pmaps = append(pmaps, pmap)
	}
	ctx.JSON(200, newHttpResp(100000, "成功获取代理列表", pmaps))
	return
}
