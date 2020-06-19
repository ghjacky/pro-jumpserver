package utils

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"time"
	"zeus/common"
)

type AccessToken struct {
	RequestTime int64    `json:"request_time"`
	User        string   `json:"user"`
	Groups      []string `json:"groups"`
	ServiceName string   `json:"service_name"`
}

var secret interface{} = []byte("D&023u@981jwoIie_!@#*s;lij!poW2ireJLAn3)-")

const TokenExpire = 60 * time.Second
const TokenNameInHeader = "Access-Token"

// 检测token是否有效、过期、字段信息等
func (at *AccessToken) ValidateToken(ctx *gin.Context, tokenString string) bool {
	if token, err := jwt.Parse(tokenString, func(token *jwt.Token) (i interface{}, e error) {
		return secret, nil
	}); err != nil {
		// 无法解析token
		common.Log.Errorf("Couldn't parse token: %s", tokenString)
		return false
	} else {
		if claim, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// 验证时间戳
			if time.Now().Unix()-int64(claim["request_time"].(float64)) < int64(TokenExpire/time.Second) {
				ctx.Set("user", "chaos")
				ctx.Set("groups", "chaos_root")
				return true
			} else {
				return false
			}
			//
		} else {
			// token无效
			common.Log.Errorln("Invalidate Token")
			return false
		}
	}
}

// 生成token返回tokenString用于设置http header
func (at *AccessToken) GenerateToken() string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"user":         at.User,
		"request_time": at.RequestTime,
		"groups":       at.Groups,
		"service_name": at.ServiceName,
	})
	if tokenString, err := token.SignedString(secret); err != nil {
		common.Log.Errorf("Token签名失败", err)
		return ""
	} else {
		return tokenString
	}
}
