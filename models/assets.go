package models

import "github.com/jinzhu/gorm"

const (
	AssetTypeSSH   = "ssh"
	AssetTypeMysql = "mysql"
	AssetTypeRedis = "redis"
)

type Asset struct {
	gorm.Model
	Type   string `gorm:"type:varchar(16); not null" json:"type"`
	Tag    string `gorm:"type:varchar(255); not null" json:"tag"`
	Expire int64  `json:"expire"`
	Server Server `json:"server"`
	User   []User
}

type Server struct {
	gorm.Model
	Type     string `gorm:"type:varchar(16); not null" json:"type"`
	Hostname string `gorm:"type:varchar(64); not null" json:"hostname"`
	IP       string `gorm:"type:varchar(32); not null" json:"ip"`
	IDC      string `gorm:"type:varchar(8); not null" json:"idc"`
	Port     int16  `gorm:"not null" json:"port"`
}

func (a *Asset) FetchList(args map[string]interface{}) (ms []IModel) {
	return
}
func (a *Asset) GetInfo(...interface{}) (err error) {
	return
}
func (a *Asset) Update() (err error) {
	return
}
func (a *Asset) Patch(...interface{}) (err error) {
	return
}
func (a *Asset) Add() (err error) {
	return
}
