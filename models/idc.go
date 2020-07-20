package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"zeus/common"
)

type SIDC struct {
	Name  string `json:"idc" gorm:"column:name;primary_key"`
	Proxy SProxy `json:"proxy" gorm:"foreignkey:idc;association_foreignkey:name"`
}
type SIDCS []*SIDC
type SProxy struct {
	gorm.Model
	IDC   string `json:"idc" gorm:"column:idc"`
	PPIP  []byte `json:"ppip" gorm:"column:ppip;not null;unique_index:p_ip_port"`
	PIP   []byte `json:"pip" gorm:"column:pip;not null;unique_index:p_ip_port"`
	PPORT uint16 `json:"pport" gorm:"column:pport;not null;unique_index:p_ip_port"`
}
type SProxies []*SProxy

func (*SIDC) TableName() string {
	return "idcs"
}

func (*SProxy) TableName() string {
	return "proxies"
}

func (idc *SIDC) UpdateOrAdd() error {
	return common.Mysql.Debug().Save(idc).Error
}

func (idc *SIDC) Delete() error {
	return common.Mysql.Debug().Delete(idc).Error
}

func (idc *SIDC) Get() error {
	return common.Mysql.First(idc).Error
}

func (idcs *SIDCS) FetchAll() error {
	return common.Mysql.Find(idcs).Error
}

func (p *SProxy) Add() error {
	if len(p.PPIP) != 4 || len(p.PIP) != 4 {
		return fmt.Errorf("proxy ip length error")
	}
	if p.PPORT <= 0 || p.PPORT > 65535 {
		return fmt.Errorf("proxy port out of bound")
	}
	return common.Mysql.Debug().Create(p).Error
}

func (p *SProxy) Delete() error {
	return common.Mysql.Debug().Delete(p).Error
}

func (ps *SProxies) FetchAll() error {
	return common.Mysql.Find(ps).Error
}

func (p *SProxy) isEmpty() bool {
	return len(p.PPIP) == 0
}

func (idc *SIDC) NeedProxy() (bool, error) {
	var proxy = SProxy{}
	if err := common.Mysql.Model(idc).Related(&proxy, "Proxy").Error; err != nil || proxy.isEmpty() {
		return false, err
	}
	idc.Proxy = proxy
	return true, nil
}
