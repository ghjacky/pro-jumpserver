package utils

import (
	"github.com/dgrijalva/jwt-go"
	"zeus/common"
)

// 加密
var encsalt interface{} = []byte("!EY*&@hbwefr7y2%qw3p;")

func Enc(src string) string {
	dest := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"src": src,
	})
	if destString, err := dest.SignedString(encsalt); err != nil {
		common.Log.Errorf("加密失败：%s", err.Error())
		return ""
	} else {
		return destString
	}
}

// 解密
func Dec(dest string) string {
	if token, err := jwt.Parse(dest, func(token *jwt.Token) (i interface{}, e error) {
		return encsalt, nil
	}); err != nil {
		return ""
	} else {
		if claim, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			return claim["src"].(string)
		} else {
			return ""
		}
	}
}
