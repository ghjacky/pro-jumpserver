package models

import "github.com/jinzhu/gorm"

type Server struct {
	gorm.Model
	PermissionID uint   `json:"permission_id"`
	Type         string `gorm:"type:varchar(16); not null; default:'ssh'" json:"type"` // default: ssh
	Hostname     string `gorm:"type:varchar(64); not null" json:"hostname"`
	IP           string `gorm:"type:varchar(32); not null" json:"ip"`
	IDC          string `gorm:"column:idc; type:varchar(32); not null" json:"idc"`
	Port         uint16 `gorm:"not null; default:22" json:"port"`
}

type Servers []*Server
