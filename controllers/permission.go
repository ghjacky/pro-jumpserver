package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"zeus/models"
	"zeus/modules/users"
)

func SetUserPerm(ctx *gin.Context) {

}

func GetUserPerm(ctx *gin.Context) {

}

/*
	Params: username
			assets
*/
func AddUserPerm(ctx *gin.Context) {
	username := ctx.Param("username")
	if len(username) == 0 {
		ctx.JSON(200, newHttpResp(100001, "参数错误: username为空", nil))
		return
	}
	var u = models.User{Username: username}
	var asts = models.Assets{}
	if err := ctx.BindJSON(&asts); err != nil {
		ctx.JSON(200, newHttpResp(100001, fmt.Sprintf("参数错误: %s", err.Error()), nil))
		return
	}
	if err := users.AddPermissionAssets(&u, &asts); err != nil {
		ctx.JSON(200, newHttpResp(100102, fmt.Sprintf("资源权限绑定失败：%s", err.Error()), nil))
		return
	}
	ctx.JSON(200, newHttpResp(100000, "资源权限绑定成功", map[string]interface{}{"user": u, "assets": asts}))
	return
}
