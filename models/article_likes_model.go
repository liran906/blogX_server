package models

import "time"

// ArticleLikesModel 文章点赞表
type ArticleLikesModel struct {
	UserID    uint      `gorm:"primaryKey" json:"userID"`
	ArticleID uint      `gorm:"primaryKey" json:"articleID"`
	CreatedAt time.Time `json:"createdAt"`

	// FK
	UserModel    UserModel    `gorm:"foreignKey:UserID;references:ID" json:"-"`
	ArticleModel ArticleModel `gorm:"foreignKey:ArticleID;references:ID" json:"-"`
}
