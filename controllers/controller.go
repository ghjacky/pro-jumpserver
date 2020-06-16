package controllers

type SHttpResp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func newHttpResp(code int, message string, data interface{}) SHttpResp {
	return SHttpResp{Code: code, Message: message, Data: data}
}
