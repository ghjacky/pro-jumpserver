package models

import (
	"zeus/common"
)

type User struct {
	Username string `gorm:"primary_key; not null" json:"username"`
	IsValid  bool   `json:"is_valid"`
	IsActive bool   `json:"is_active"`
	Assets   Assets `gorm:"many2many:user_asset" json:"asset"`
}

//type Permission struct {
//	Username 		string
//	AssetID 		string
//}

func (u *User) FetchList(args map[string]interface{}) (ms []IModel) {
	return append(ms, &User{Username: "myguo", IsValid: true, IsActive: false})
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
