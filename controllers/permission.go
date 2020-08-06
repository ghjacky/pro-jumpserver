package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"zeus/models"
	"zeus/modules/webserver/permission"
)

func SetUserPermissions(ctx *gin.Context) {

}

/*
	Params: username
*/
func GetUserPermissions(ctx *gin.Context) {
	username := ctx.Param("username")
	if len(username) == 0 {
		ctx.JSON(200, newHttpResp(100001, "参数错误: username为空", nil))
		return
	}
	var u = models.User{Username: username}
	if err := permission.FetchPermissions(&u); err != nil {
		ctx.JSON(200, newHttpResp(100102, fmt.Sprintf("获取用户: %s 的权限资源失败: %s", username, err.Error()), nil))
		return
	}
	ctx.JSON(200, newHttpResp(100000, "成功获取用户权限资源", u))
	return
}

/*
	Params: username
			assets
*/
func AddUserPermissions(ctx *gin.Context) {
	username := ctx.Param("username")
	if len(username) == 0 {
		ctx.JSON(200, newHttpResp(100001, "参数错误：用户名为空", nil))
		return
	}
	var asts = models.Permissions{}
	if err := ctx.BindJSON(&asts); err != nil {
		ctx.JSON(200, newHttpResp(100001, fmt.Sprintf("参数错误: %s", err.Error()), nil))
		return
	}
	if err := permission.AddPermissions(username, &asts); err != nil {
		ctx.JSON(200, newHttpResp(100102, fmt.Sprintf("资源权限绑定失败：%s", err.Error()), nil))
		return
	}
	ctx.JSON(200, newHttpResp(100000, "资源权限绑定成功", map[string]interface{}{"username": username, "assets": asts}))
	return
}

func FetchPermissions(ctx *gin.Context) {
	query := models.Query{}
	if err := ctx.BindQuery(&query); err != nil {
		ctx.JSON(200, newHttpResp(100001, fmt.Sprintf("参数错误：%s", err.Error()), nil))
		return
	}
	perms, total, err := permission.FetchAllPermissions(query)
	if err != nil {
		ctx.JSON(200, newHttpResp(100002, fmt.Sprintf("数据库错误：%s", err.Error()), nil))
		return
	}
	ctx.JSON(200, newHttpResp(100000, "成功获取权限资产", map[string]interface{}{"total": total, "perms": perms}))
	return
}

func DeletePermission(ctx *gin.Context) {
	pid, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(200, newHttpResp(100001, fmt.Sprintf("参数错误：%s", err.Error()), nil))
		return
	}
	perm := &models.Permission{}
	perm.ID = uint(pid)
	if err := permission.DeletePermission(perm); err != nil {
		ctx.JSON(200, newHttpResp(100002, fmt.Sprintf("数据库错误：%s", err.Error()), nil))
		return
	}
	ctx.JSON(200, newHttpResp(100000, "删除成功", perm))
	return
}
