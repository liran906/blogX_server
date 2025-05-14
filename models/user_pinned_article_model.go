package models

import "time"

type UserPinnedArticleModel struct {
	UserID    uint      `gorm:"primaryKey" json:"userID"`
	ArticleID uint      `gorm:"primaryKey" json:"articleID"`
	CreatedAt time.Time `json:"createdAt"`

	// FK
	UserModel    UserModel    `gorm:"foreignKey:UserID;references:ID" json:"-"`
	ArticleModel ArticleModel `gorm:"foreignKey:ArticleID;references:ID" json:"-"`
}
