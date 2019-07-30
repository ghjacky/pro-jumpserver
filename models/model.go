package models

type models []IModel

type IModel interface {
	FetchList() models
	GetInfo(...interface{}) error
	Update() error
	Patch(...interface{}) error
}
