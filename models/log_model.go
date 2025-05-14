package models

type LogModel struct {
	Model
	LogType uint8  `gorm:"not null" json:"logType"` // 日志类型
	Title   string `gorm:"size:128; not null" json:"title"`
	Content string `gorm:"not null" json:"content"`
	Level   uint8  `gorm:"not null" json:"level"`
	UserID  *uint  `json:"userID"`
	IP      string `gorm:"size:32; not null" json:"ip"`
	Address string `gorm:"size:64; not null" json:"address"`
	IsRead  bool   `gorm:"not null; default:false" json:"isRead"`

	// FK
	UserModel UserModel `gorm:"foreignKey:UserID;references:ID" json:"-"`
}
