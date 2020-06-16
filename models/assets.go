package models

import (
	"github.com/jinzhu/gorm"
	"zeus/common"
)

const (
	AssetTypeSSH   = "ssh"
	AssetTypeMysql = "mysql"
	AssetTypeRedis = "redis"
)

type Asset struct {
	gorm.Model
	Type    string  `gorm:"type:varchar(16); not null" json:"type"`
	Tag     string  `gorm:"type:varchar(255)" json:"tag"` // 一个asset对应一个tag，一个tag对应多个server
	Expire  int64   `json:"expire"`
	Servers Servers `json:"servers" gorm:"many2many:asset_server"` // 如果preload有关联的servers则说明直接绑定的主机，可直接获取使用即可，如果没有关联的servers，则根据tag获取servers列表
}

type Assets []*Asset

type Server struct {
	gorm.Model
	Type     string `gorm:"type:varchar(16); not null" json:"type"`
	Hostname string `gorm:"type:varchar(64); not null" json:"hostname"`
	IP       string `gorm:"type:varchar(32); not null" json:"ip"`
	IDC      string `gorm:"type:varchar(8); not null" json:"idc"`
	Port     int16  `gorm:"not null" json:"port"`
}

type Servers []*Server

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
	return common.Mysql.Debug().Save(a).Error
}
