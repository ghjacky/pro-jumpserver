package models

//type IModel interface {
//	FetchList(map[string]interface{}) []IModel
//	GetInfo(...interface{}) error
//	Update() error
//	Patch(...interface{}) error
//	Add() error
//}

type Query struct {
	Dimension string `form:"dimension"`
	Search    string `form:"search"`
	Page      int    `form:"page"`
	Limit     int    `form:"limit"`
	Sort      string `form:"sort"`
}
