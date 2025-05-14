package models

type CategoryModel struct {
	Model
	Name   string `gorm:"size:32; not null" json:"name"`
	UserID uint   `gorm:"not null" json:"userID"` // 创建人

	// FK
	UserModel UserModel `gorm:"foreignKey:UserID;references:ID" json:"-"`
}
