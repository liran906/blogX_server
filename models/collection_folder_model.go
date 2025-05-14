package models

type CollectionFolderModel struct {
	Model
	Title        string `gorm:"size:128; not null" json:"title"`
	Abstract     string `gorm:"size:256" json:"abstract"`
	CoverURL     string `gorm:"size:256" json:"coverURL"`
	ArticleCount int    `gorm:"not null" json:"articleCount"`
	UserID       uint   `gorm:"not null" json:"userID"`

	// FK
	UserModel UserModel `gorm:"foreignKey:UserID;references:ID" json:"-"`
}
