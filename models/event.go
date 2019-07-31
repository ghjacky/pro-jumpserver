package models

// 定义event
type Event struct {
	SessionID    string `gorm:"type:varchar(128); not null" json:"session_id"`
	SubSessionID string `gorm:"type:varchar(128); not null" json:"sub_session_id"`
	Type         string `gorm:"type:varchar(64);not null" json:"type"`
	Err          string `gorm:"type:varchar(255)" json:"err"`
	User         string `gorm:"type:varchar(32); not null" json:"user"`
	Timestamp    int64  `gorm:"not null" json:"timestamp"`
	ClientIP     string `gorm:"type:varchar(32);not null" json:"client_ip"`
	ServerIP     string `gorm:"type:varchar(32);not null" json:"server_ip"`
	SrcFile      string `gorm:"type:varchar(255)" json:"src_file"`
	DestFile     string `gorm:"type:varchar(255)" json:"dest_file"`
	Bin          string `gorm:"type:varchar(32)" json:"bin"`
	Command      string `gorm:"type:varchar(255)" json:"command"`
	Data         []byte `gorm:"type:varbinary" json:"data"`
}
