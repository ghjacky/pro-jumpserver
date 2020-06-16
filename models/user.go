package models

import (
	"time"
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
	return common.Mysql.Model(u).Preload("Assets", "username = ?", u.Username).Find(u).Error
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

func (u *User) FetchPermissionAssets(idc string) (ms []IModel) {
	// 根据tag从服务树拉取主机列表，并整理生成assets列表，返回
	//
	s := Server{}
	s.ID = 1
	s.IP = "172.16.244.28"
	s.Hostname = "dev_server"
	s.IDC = "天津"
	s.Type = "ssh"
	s.Port = 22
	a := Asset{}
	a.ID = 1
	a.Type = "ssh"
	a.Servers = Servers{&s}
	a.Tag = ""
	a.Expire = time.Now().UnixNano() + int64(3*time.Hour)
	return append(ms, &a)
}
