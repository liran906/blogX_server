package models

type CollectionFolderModel struct {
	Model
	UserID       uint   `gorm:"not null;uniqueIndex:idx_uniq_collection_folder;" json:"userID"`
	Title        string `gorm:"not null;uniqueIndex:idx_uniq_collection_folder;size:128; not null" json:"title"`
	Abstract     string `gorm:"size:256" json:"abstract"`
	CoverURL     string `gorm:"size:256" json:"coverURL"`
	ArticleCount int    `gorm:"not null" json:"articleCount"`
	IsDefault    bool   `gorm:"default:false" json:"isDefault"` // 是否是默认收藏夹

	// FK
	UserModel       UserModel       `gorm:"foreignKey:UserID;references:ID" json:"-"`
	UserConfigModel UserConfigModel `gorm:"foreignKey:UserID;references:UserID" json:"-"`
}
