// Path: ./models/log_model.go

package models

import "blogX_server/models/enum"

type LogModel struct {
	Model
	LogType     enum.LogType      `gorm:"not null" json:"logType"` // 日志类型
	Title       string            `gorm:"size:128; not null" json:"title"`
	Content     string            `gorm:"not null" json:"content"`
	Level       enum.LogLevelType `gorm:"not null" json:"level"`
	UserID      uint              `json:"userID"`
	Username    string            `gorm:"size:32; not null" json:"username"` // 登录日志的用户名
	Password    string            `gorm:"size:32; not null" json:"password"` // 登录日志的密码
	IP          string            `gorm:"size:32; not null" json:"ip"`
	Address     string            `gorm:"size:64; not null" json:"address"`
	IsRead      bool              `gorm:"not null; default:false" json:"isRead"`
	LoginStatus bool              `gorm:"not null; default:false" json:"loginStatus"` // 登录状态
	LoginType   enum.LoginType    `gorm:"not null" json:"loginType"`                  // 登录的类型
	UA          string            `gorm:"size:256; not null" json:"ua"`               // 登录设备
	ServiceName string            `gorm:"size:32" json:"serviceName"`

	// FK
	UserModel UserModel `gorm:"foreignKey:UserID;references:ID" json:"-"`
}
