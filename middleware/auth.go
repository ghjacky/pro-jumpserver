package middleware

import (
	"github.com/gin-gonic/gin"
	"regexp"
	"zeus/utils"
)

var Token = utils.AccessToken{}

func CheckToken() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if !isMustApi(ctx) {
			return
		}
		tokenString := ctx.GetHeader(utils.TokenNameInHeader)
		if b := Token.ValidateToken(ctx, tokenString); b {
			ctx.Next()
			return
		} else {
			ctx.JSON(200, map[string]interface{}{"code": 100001, "message": "access token无效"})
			ctx.Abort()
			return
		}
	}
}

// 定义无需登陆检测的接口
func isMustApi(ctx *gin.Context) bool {
	return true

}

func ignoreMatchErr(pattern, str string) bool {
	match, _ := regexp.MatchString(pattern, str)
	return match
}
