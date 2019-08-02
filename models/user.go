package models

import (
	"time"
	"zeus/common"
)

type User struct {
	Username string  `gorm:"primary_key; not null" json:"username"`
	IsValid  bool    `json:"is_valid"`
	IsActive bool    `json:"is_active"`
	Asset    []Asset `gorm:"many2many:user_asset" json:"asset"`
}

//type Permission struct {
//	Username 		string
//	AssetID 		string
//}

func (u *User) FetchList(args map[string]interface{}) (ms []IModel) {
	return append(ms, &User{Username: "myguo", IsValid: true, IsActive: false})
}
func (u *User) GetInfo(...interface{}) (err error) {
	if err = common.Mysql.Model(u).Preload("Asset", "username = ?", u.Username).Find(u).Error; err != nil {
		common.Log.Errorf("couldn't get asset info of user: %s", u.Username)
	}
	return
}
func (u *User) Update() (err error) {
	if err = common.Mysql.Save(u).Error; err != nil {
		common.Log.Error("Couldn't update user info")
	}
	return
}
func (u *User) Patch(...interface{}) (err error) {
	return
}
func (u *User) Add() (err error) {
	return
}

func (u *User) FetchPermissionAssets(idc string) (ms []IModel) {
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
	a.Server = s
	a.Tag = ""
	a.Expire = time.Now().UnixNano() + int64(3*time.Hour)
	return append(ms, &a)
}
