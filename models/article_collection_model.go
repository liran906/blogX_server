package models

import "time"

type ArticleCollectionModel struct {
	// user article collectionFolder 三个字段作为联合主键
	// 一篇文章可以多个收藏夹
	UserID             uint      `gorm:"primaryKey" json:"userID"`
	ArticleID          uint      `gorm:"primaryKey" json:"articleID"`
	CollectionFolderID uint      `gorm:"primaryKey" json:"collectionFolderID"`
	CreatedAt          time.Time `json:"createdAt"`

	// FK
	UserModel             UserModel             `gorm:"foreignKey:UserID;references:ID" json:"-"`
	ArticleModel          ArticleModel          `gorm:"foreignKey:ArticleID;references:ID" json:"-"`
	CollectionFolderModel CollectionFolderModel `gorm:"foreignKey:CollectionFolderID;references:ID" json:"-"`
}
