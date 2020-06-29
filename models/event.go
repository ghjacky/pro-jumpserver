package models

import (
	"fmt"
	"github.com/google/uuid"
	"strings"
	"time"
	"zeus/common"
)

// 定义event
type Event struct {
	ID        uuid.UUID `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
	SessionID string     `gorm:"type:varchar(128); not null" json:"session_id"`
	Type      string     `gorm:"type:varchar(64);not null" json:"type"`
	Err       string     `gorm:"type:varchar(255)" json:"err"`
	Timestamp int64      `gorm:"-" json:"timestamp"`
	ClientIP  string     `gorm:"type:varchar(32);not null" json:"client_ip"`
	ServerIP  string     `gorm:"type:varchar(32);not null" json:"server_ip"`
	SrcFile   string     `gorm:"-" json:"src_file"`
	DestFile  string     `gorm:"-" json:"dest_file"`
	Bin       string     `gorm:"-" json:"bin"`
	Command   string     `gorm:"-" json:"command"`
	Data      []byte     `gorm:"-" json:"data"`
	Username  string     `gorm:"type:varchar(255);not null" json:"username"`
}

type Events []*Event

func (es *Events) GetEvents(query Query) (total int, err error) {
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
	if err = common.Mysql.Model(Event{}).Where(whereClause).Count(&total).Error; err != nil {
		return 0, err
	}
	if err = common.Mysql.Where(whereClause).Order(fmt.Sprintf("%s %s", orderBy, orderType)).Offset(offset).Limit(limit).Find(es).Error; err != nil {
		return 0, err
	}
	return
}
