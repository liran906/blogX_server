package models

type BannerModel struct {
	Model
	URL  string `gorm:"size:256; not null" json:"url"`  // 图片链接
	Href string `gorm:"size:256; not null" json:"href"` // 跳转链接
}
