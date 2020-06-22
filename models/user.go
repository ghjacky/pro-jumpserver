package models

import (
	"zeus/common"
)

type User struct {
	Username string `gorm:"primary_key; not null" json:"username"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Valid    string `json:"valid" gorm:"default:'否'"`   // 账户可用状态：是、否
	Active   string `json:"active" gorm:"default:'下线'"` // 账户登陆状态：在线、下线
	Assets   Assets `gorm:"many2many:user_asset" json:"asset"`
}

type Users []*User

//type Permission struct {
//	Username 		string
//	AssetID 		string
//}

const (
	UserValidYes  = "是"
	UserValidNo   = "否"
	UserActiveYes = "在线"
	UserActiveNo  = "下线"
)

func (u *User) FetchList(args map[string]interface{}) (ms []IModel) {
	return append(ms, &User{Username: "myguo", Valid: UserValidYes, Active: UserActiveNo})
}
func (u *User) GetInfo(...interface{}) (err error) {
	return common.Mysql.Preload("Assets").Find(u).Error
}
func (u *User) Update() (err error) {
	return common.Mysql.Debug().Save(u).Error
}
func (u *User) ReplaceAssets() error {
	return common.Mysql.Model(u).Association("Assets").Replace(u.Assets).Error
}
func (u *User) Patch(...interface{}) (err error) {
	return
}
func (u *User) Add() (err error) {
	return common.Mysql.Debug().Create(u).Error
}
