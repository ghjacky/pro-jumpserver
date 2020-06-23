package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"zeus/models"
)

func FetchUserList(ctx *gin.Context) {
	var query models.Query
	if err := ctx.BindQuery(&query); err != nil {
		ctx.JSON(200, newHttpResp(100001, fmt.Sprintf("参数错误: %s", err.Error()), nil))
		return
	}
	user := models.User{}
	total, users, err := user.FetchList(query)
	if err != nil {
		ctx.JSON(200, newHttpResp(100202, fmt.Sprintf("获取用户列表出错：%s", err.Error()), nil))
		return
	}
	ctx.JSON(200, newHttpResp(100000, "成功获取用户列表", map[string]interface{}{"total": total, "users": users}))
	return
}

func ValidUser(ctx *gin.Context) {
	username := ctx.Param("username")
	if len(username) == 0 {
		ctx.JSON(200, newHttpResp(100001, "参数错误：用户名为空", nil))
		return
	}
	valid, err := strconv.Atoi(ctx.Query("valid"))
	if err != nil {
		ctx.JSON(200, newHttpResp(100001, "参数错误", nil))
		return
	}
	user := models.User{Username: username}
	if valid == 1 {
		user.Valid = models.UserValidYes
	} else {
		user.Valid = models.UserValidNo
	}
	if err := user.SetValid(); err != nil {
		ctx.JSON(200, newHttpResp(100202, fmt.Sprintf("更改账户可用性出错：%s", err.Error()), nil))
		return
	}
	ctx.JSON(200, newHttpResp(100000, "成功更新账户可用性状态", valid))
	return
}
