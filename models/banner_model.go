package models

type BannerModel struct {
	Model
	Activated bool   `gorm:"not null; default:true" json:"activated"` // 是否展示
	URL       string `gorm:"size:256; not null" json:"url"`           // 图片链接
	Href      string `gorm:"size:256; not null" json:"href"`          // 跳转链接
}
