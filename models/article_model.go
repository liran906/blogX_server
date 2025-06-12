package models

import (
	"blogX_server/models/ctype"
	"blogX_server/models/enum"
	"blogX_server/service/text_service"
	_ "embed"
	"gorm.io/gorm"
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
	PinnedByUser   bool               `gorm:"not null; default:false" json:"pinnedByUser"`
	PinnedByAdmin  bool               `gorm:"not null; default:false" json:"pinnedByAdmin"`

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

// AfterCreate 在创建后把文章内容按照 md 格式拆分为小标题+文字的形式，并且入 text_model 库
func (a *ArticleModel) AfterCreate(tx *gorm.DB) (err error) {
	// 已发布的文章才进行
	if a.Status != enum.ArticleStatusPublish {
		return
	}

	_list := text_service.MDContentTransformation(a.ID, a.Title, a.Content)

	if len(_list) == 0 {
		return nil
	}

	var list []TextModel
	for _, txt := range _list {
		list = append(list, TextModel{
			ArticleID: a.ID,
			Head:      txt.Head,
			Body:      txt.Body,
		})
	}

	err = tx.Create(&list).Error
	return
}

// BeforeDelete 在删除文章之前，把对应的 text_model 删除
func (a *ArticleModel) BeforeDelete(tx *gorm.DB) (err error) {
	err = tx.Where("article_id = ?", a.ID).Delete(&TextModel{}).Error
	return
}

// AfterUpdate 在删除文章后，把对应的 text_model 删除再重建
// 也可以到对应 api 去判断文章（标题 正文 状态）有没有变化，有变化才执行重构
func (a *ArticleModel) AfterUpdate(tx *gorm.DB) (err error) {
	err = a.BeforeDelete(tx)
	if err != nil {
		return
	}
	return a.AfterCreate(tx)
}
