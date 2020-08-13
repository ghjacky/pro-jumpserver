package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"strings"
	"time"
	"zeus/common"
)

const (
	ServerTypeSSH        = "ssh"
	ServerTypeMysql      = "mysql"
	ServerTypeRedis      = "redis"
	PermissionTypeTag    = 0
	PermissionTypeServer = 1
)

type Permission struct {
	gorm.Model
	Username string    `json:"username"`
	Type     int8      `gorm:"not null" json:"type"`         // 权限类型：0: "tag" or 1: "host"
	Tag      string    `gorm:"type:varchar(255)" json:"tag"` // 一个asset对应一个tag，一个tag对应多个server
	Period   uint16    `json:"period"`
	Expire   time.Time `json:"expire"`
	Sudo     uint8     `json:"sudo" gorm:"default:0"` // 0: 不添加sudo权限 1: 添加sudo权限， 默认：0
	Servers  Servers   `json:"servers"`               // 如果preload有关联的servers则说明直接绑定的主机，可直接获取使用即可，如果没有关联的servers，则根据tag获取servers列表
}

type Permissions []*Permission

func (p *Permission) BeforeCreate(db *gorm.DB) {
	if p.Period == 0 {
		p.Expire = time.Now().Add(-1 * 24 * time.Hour)
	} else {
		p.Expire = time.Now().Add(time.Duration(p.Period) * 24 * time.Hour)
	}
}

func (p *Permission) AfterFind(db *gorm.DB) {
	if p.IsExpire() {
		db.Delete(p)
		return
	}
	db.Model(p).Association("Servers").Find(&p.Servers)
}

func (p *Permission) BeforeDelete(db *gorm.DB) {
	db.Where("permission_id = ?", p.ID).Delete(p.Servers)
}

func (ps *Permissions) FetchList(query Query) (total int, err error) {
	total = 0
	var whereClause, orderClause string
	if len(query.Dimension) != 0 {
		whereClause = fmt.Sprintf("%s like '%%%s%%'", query.Dimension, query.Search)
	}
	if len(query.Sort) != 0 && strings.HasPrefix(query.Sort, "+") {
		orderClause = fmt.Sprintf("%s asc", strings.TrimPrefix(query.Sort, "+"))
	} else if len(query.Sort) != 0 && strings.HasPrefix(query.Sort, "-") {
		orderClause = fmt.Sprintf("%s desc", strings.TrimPrefix(query.Sort, "-"))
	}
	common.Mysql.Model(ps).Where(whereClause).Count(&total)
	return total, common.Mysql.Where(whereClause).Preload("Servers").Order(orderClause).Offset((query.Page - 1) * query.Limit).Limit(query.Limit).Find(ps).Error
}

func (p *Permission) GetInfo() (err error) {
	return common.Mysql.Preload("Servers").First(p).Error
}

func (p *Permission) Update() (err error) {
	return
}

func (p *Permission) Patch(...interface{}) (err error) {
	return
}

func (p *Permission) Add() (err error) {
	return common.Mysql.Debug().Create(p).Error
}

func (p *Permission) Delete() (err error) {
	return common.Mysql.Debug().Delete(p).Error
}

func (p *Permission) IsExpire() bool {
	return p.Expire.After(p.CreatedAt) && p.Expire.Before(time.Now())
}