// Path: ./models/text_model.go

package models

import (
	_ "embed"
)

type TextModel struct {
	Model
	ArticleID uint   `gorm:"not null" json:"articleID"`
	Head      string `json:"head"`
	Body      string `json:"body"`

	// FK
	//ArticleModel ArticleModel `gorm:"foreignKey:ArticleID;references:ID" json:"-"`
}

// `go:embed`用于在 编译时 把文件内容打包进 Go 二进制文件 中。这样就不需要在运行时再去加载外部文件了。

//go:embed mappings/text_mapping.json
var textMapping string

func (TextModel) Mapping() string {
	return textMapping
}

// GetIndex 获取索引名字（index 就像 mysql 中的 table name 一样）
func (TextModel) GetIndex() string {
	return "text_index"
}
