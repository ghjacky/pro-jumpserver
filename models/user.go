package models

import (
	"fmt"
	"strings"
	"zeus/common"
)

type User struct {
	Username    string      `gorm:"primary_key; not null" json:"username"`
	Nickname    string      `json:"nickname"`
	Email       string      `json:"email"`
	Valid       string      `json:"valid" gorm:"default:'否'"`   // 账户可用状态：是、否
	Active      string      `json:"active" gorm:"default:'下线'"` // 账户登陆状态：在线、下线
	Permissions Permissions `gorm:"foreignkey:Username" json:"permissions"`
	Events      Events      `json:"events" gorm:"foreignkey:Username"`
}

type Users []*User

const (
	UserValidYes  = "是"
	UserValidNo   = "否"
	UserActiveYes = "在线"
	UserActiveNo  = "下线"
)

func (u *User) FetchList(query Query) (total int, users Users, err error) {
	var offset = 0
	var limit = 9999
	if query.Page != 0 && query.Limit != 0 {
		limit = query.Limit
		offset = (query.Page - 1) * query.Limit
	}
	whereClause := fmt.Sprintf("%s like '%%%s%%'", query.Dimension, query.Search)
	var orderBy = "username"
	var orderType = "asc"
	if strings.HasPrefix(query.Sort, "+") {
		orderBy = strings.TrimPrefix(query.Sort, "+")
		orderType = "asc"
	} else if strings.HasPrefix(query.Sort, "-") {
		orderBy = strings.TrimPrefix(query.Sort, "-")
		orderType = "desc"
	}
	if err = common.Mysql.Model(u).Where(whereClause).Count(&total).Error; err != nil {
		return
	}
	if err = common.Mysql.Where(whereClause).Order("active desc").Order(fmt.Sprintf("%s %s", orderBy, orderType)).Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return
	}
	return
}
func (u *User) SetValid() (err error) {
	return common.Mysql.Model(u).UpdateColumns(User{Valid: u.Valid}).Error
}
func (u *User) GetInfo(...interface{}) (err error) {
	return common.Mysql.Preload("Assets").Find(u).Error
}
func (u *User) Update() (err error) {
	return common.Mysql.Debug().Save(u).Error
}

func (u *User) Patch(...interface{}) (err error) {
	return
}
func (u *User) Add() (err error) {
	return common.Mysql.Debug().Create(u).Error
}
