package models

type ArticleCollectionModel struct {
	// user article collectionFolder 三个字段作为联合 CK
	// 一篇文章可以多个收藏夹
	Model
	UserID             uint `gorm:"not null;uniqueIndex:idx_uniq_article_collection" json:"userID"`
	ArticleID          uint `gorm:"not null;uniqueIndex:idx_uniq_article_collection" json:"articleID"`
	CollectionFolderID uint `gorm:"not null;uniqueIndex:idx_uniq_article_collection" json:"collectionFolderID"`

	// FK
	UserModel             UserModel             `gorm:"foreignKey:UserID;references:ID" json:"-"`
	ArticleModel          ArticleModel          `gorm:"foreignKey:ArticleID;references:ID" json:"-"`
	CollectionFolderModel CollectionFolderModel `gorm:"foreignKey:CollectionFolderID;references:ID" json:"-"`
}
