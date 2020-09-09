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

func (p *Permission) BeforeCreate(db *gorm.DB) error {
	if p.Period == 0 {
		p.Expire = time.Now().Add(-1 * 24 * time.Hour)
	} else {
		p.Expire = time.Now().Add(time.Duration(p.Period) * 24 * time.Hour)
	}
	// 判断该权限下的tag或者主机是否已存在于该用户的其他权限条目下
	if p.Type == 1 {
		perms := new(Permissions)
		if e := perms.GetHostsPermByUsername(p.Username, db); e != nil && !gorm.IsRecordNotFoundError(e) {
			return e
		} else if gorm.IsRecordNotFoundError(e) {
			return nil
		}
		servers := perms.GetPermedHosts()
		var serverMap = map[string]bool{}
		if len(servers) > 0 {
			for _, s := range servers {
				serverMap[s.IDC+"-"+s.IP] = true
			}
		}
		for _, s := range p.Servers {
			if v, ok := serverMap[s.IDC+"-"+s.IP]; ok && v {
				return fmt.Errorf("server (%s-%s) already permed", s.IDC, s.IP)
			}
		}
	} else if p.Type == 0 {
		perm := new(Permission)
		if e := perm.GetPermByUsernameAndTag(p.Username, p.Tag, db); e != nil && !gorm.IsRecordNotFoundError(e) {
			return e
		} else if gorm.IsRecordNotFoundError(e) {
			return nil
		} else {
			return fmt.Errorf("tag %s already permed", p.Tag)
		}
	}
	return nil
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

func (p *Permission) Update(db *gorm.DB) (err error) {
	var _p = new(Permission)
	_p.ID = p.ID
	if err := _p.Delete(db); err != nil && !gorm.IsRecordNotFoundError(err) {
		return err
	}
	p.ID = 0
	if err := p.Add(db); err != nil {
		return err
	}
	return nil
}

func (p *Permission) Patch(...interface{}) (err error) {
	return
}

func (p *Permission) Add(db *gorm.DB) (err error) {
	return db.Debug().Create(p).Error
}

func (p *Permission) Delete(db *gorm.DB) (err error) {
	return db.Debug().Unscoped().Delete(p).Error
}

func (p *Permission) IsExpire() bool {
	return p.Expire.After(p.CreatedAt) && p.Expire.Before(time.Now())
}

func (ps *Permissions) GetHostsPermByUsername(username string, db *gorm.DB) (err error) {
	var t = PermissionTypeServer
	return db.Model(&Permission{}).Where("username = ?", username).Where("type = ?", t).Find(ps).Error
}

func (ps *Permissions) GetPermedHosts() (ss Servers) {
	for _, p := range *ps {
		ss = append(ss, p.Servers...)
	}
	return
}

func (p *Permission) GetPermByUsernameAndTag(username, tag string, db *gorm.DB) (err error) {
	var t = PermissionTypeTag
	return db.Model(&Permission{}).Where("username = ?", username).Where("tag = ?", tag).Where("type = ?", t).First(p).Error
}
