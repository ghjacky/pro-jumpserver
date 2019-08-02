package models

type IModel interface {
	FetchList(map[string]interface{}) []IModel
	GetInfo(...interface{}) error
	Update() error
	Patch(...interface{}) error
	Add() error
}
