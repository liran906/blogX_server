// Path: ./models/text_model.go

package models

type TextModel struct {
	Model
	ArticleID uint   `gorm:"not null" json:"articleID"`
	Head      string `json:"head"`
	Body      string `json:"body"`

	// FK
	//ArticleModel ArticleModel `gorm:"foreignKey:ArticleID;references:ID" json:"-"`
}
