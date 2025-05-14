package models

type UserArticleHistoryModel struct {
	Model
	UserID     uint  `gorm:"not null" json:"userID"`
	ArticleID  uint  `gorm:"not null" json:"articleID"`
	Percentage uint8 `gorm:"not null" json:"percentage"`

	// FK
	UserModel    UserModel    `gorm:"foreignKey:UserID;references:ID" json:"-"`
	ArticleModel ArticleModel `gorm:"foreignKey:ArticleID;references:ID" json:"-"`
}
