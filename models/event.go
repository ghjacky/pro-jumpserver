package models

import (
	"github.com/google/uuid"
	"time"
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
	User      string     `gorm:"type:varchar(32); not null" json:"user"`
	Timestamp int64      `gorm:"-" json:"timestamp"`
	ClientIP  string     `gorm:"type:varchar(32);not null" json:"client_ip"`
	ServerIP  string     `gorm:"type:varchar(32);not null" json:"server_ip"`
	SrcFile   string     `gorm:"-" json:"src_file"`
	DestFile  string     `gorm:"-" json:"dest_file"`
	Bin       string     `gorm:"-" json:"bin"`
	Command   string     `gorm:"-" json:"command"`
	Data      []byte     `gorm:"-" json:"data"`
}
