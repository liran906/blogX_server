package models

import (
	"blogX_server/models/ctype"
	"blogX_server/models/enum"
	_ "embed"
)

type ArticleModel struct {
	Model
	Title          string             `gorm:"size:128; not null" json:"title"`
	Abstract       string             `gorm:"size:256" json:"abstract"`
	CoverURL       string             `gorm:"size:256" json:"coverURL"`
	Content        string             `gorm:"not null" json:"content"`
	CategoryID     *uint              `json:"categoryID"`
	Tags           ctype.List         `gorm:"type:longtext" json:"tags"`
	UserID         uint               `gorm:"not null" json:"userID"`
	Status         enum.ArticleStatus `json:"status"` // 草稿 审核中 已发布
	ReadCount      int                `gorm:"not null" json:"readCount"`
	LikeCount      int                `gorm:"not null" json:"likeCount"`
	CommentCount   int                `gorm:"not null" json:"commentCount"`
	OpenForComment bool               `gorm:"not null; default:true" json:"openForComment"`

	// FK
	UserModel UserModel `gorm:"foreignKey:UserID;references:ID" json:"-"`
}

// `go:embed`用于在 编译时 把文件内容打包进 Go 二进制文件 中。这样就不需要在运行时再去加载外部文件了。

//go:embed mappings/article_mapping.json
var articleMapping string

func (ArticleModel) Mapping() string {
	return articleMapping
}

// GetIndex 获取索引名字（index 就像 mysql 中的 table name 一样）
func (ArticleModel) GetIndex() string {
	return "article_index"
}
