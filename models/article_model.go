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
	CategoryID     *uint              `json:"categoryID"`                // 自定义分类
	Tags           ctype.List         `gorm:"type:longtext" json:"tags"` // 标签
	UserID         uint               `gorm:"not null" json:"userID"`    // 发布者
	Status         enum.ArticleStatus `json:"status"`                    // 草稿 审核中 已发布
	ReadCount      int                `gorm:"not null; default:0" json:"readCount"`
	LikeCount      int                `gorm:"not null; default:0" json:"likeCount"`
	CommentCount   int                `gorm:"not null; default:0" json:"commentCount"`
	CollectCount   int                `gorm:"not null; default:0" json:"collectCount"`
	OpenForComment bool               `gorm:"not null; default:true" json:"openForComment"`
	PinnedByUser   bool               `gorm:"not null; default:false" json:"pinnedByUser"` // 0就是没有被置顶，其他数字就是置顶顺序，1为最顶
	PinnedByAdmin  uint               `gorm:"not null; default:0" json:"pinnedByAdmin"`    // 0就是没有被置顶，其他数字就是置顶顺序，1为最顶

	// FK
	UserModel     UserModel      `gorm:"foreignKey:UserID;references:ID" json:"-"`
	CategoryModel *CategoryModel `gorm:"foreignKey:CategoryID;references:ID" json:"-"`
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
