package models

type CategoryModel struct {
	Model
	Name   string `gorm:"not null;uniqueIndex:idx_uniq_category_name;size:32" json:"name"`
	UserID uint   `gorm:"not null;uniqueIndex:idx_uniq_category_name" json:"userID"` // 创建人

	// FK
	UserModel   UserModel      `gorm:"foreignKey:UserID;references:ID" json:"-"`
	ArticleList []ArticleModel `gorm:"foreignKey:CategoryID;references:ID" json:"-"`
}
