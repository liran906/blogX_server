package models

type GlobalNotificationModel struct {
	Model
	Title   string `gorm:"size:64; not null" json:"title"`
	Content string `gorm:"size:256; not null" json:"content"`
	IconURL string `gorm:"size:256" json:"iconURL"`
	Herf    string `gorm:"size:256; not null" json:"href"` // 跳转链接
}
