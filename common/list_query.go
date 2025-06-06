// Path: ./common/list_query.go

package common

import (
	"blogX_server/global"
	"fmt"
	"gorm.io/gorm"
)

type PageInfo struct {
	Limit int    `form:"limit"`
	Page  int    `form:"page"`
	Key   string `form:"key"`
	Order string `form:"order"` // 这个 order 由前端写入，优先级高于 defaultorder
}

func (p *PageInfo) Normalize() {
	p.Page = p.GetPage()
	p.Limit = p.GetLimit()
}

// GetPage 确保 page 合理
func (p *PageInfo) GetPage() int {
	if p.Page < 1 {
		p.Page = 1
	}
	return p.Page
}

// GetLimit 确保 limit 合理
func (p *PageInfo) GetLimit() int {
	if p.Limit < 1 || p.Limit > 100 {
		p.Limit = 10
	}
	return p.Limit
}

// GetOffset 计算 offset
func (p *PageInfo) GetOffset() int {
	return (p.GetPage() - 1) * p.GetLimit()
}

// Options 搜索选项
type Options struct {
	PageInfo     PageInfo
	Likes        []string // 模糊匹配的字段
	Preloads     []string // 预加载关联数据（比如 FK）
	Where        *gorm.DB // 定制化查询 TBD
	Debug        bool     // 是否要看SQL语句
	DefaultOrder string   // 顺序，这个前端可覆盖
}

// ListQuery performs a database query with support for pagination, filtering, ordering, and preloading related data.
func ListQuery[T any](model T, options Options) (list []T, count int, err error) {
	// 基础查询
	query := global.DB.Model(model)

	// 预加载
	for _, model := range options.Preloads {
		query = query.Preload(model)
	}

	// SQL 语句显示
	if options.Debug {
		query = query.Debug()
	}

	// 精确匹配
	query = query.Where(model)

	// 模糊匹配
	if len(options.Likes) > 0 && options.PageInfo.Key != "" {
		likes := global.DB
		for i, column := range options.Likes {
			if i == 0 {
				likes = likes.Where(fmt.Sprintf("%s LIKE ?", column), "%"+options.PageInfo.Key+"%")
			} else {
				likes = likes.Or(fmt.Sprintf("%s LIKE ?", column), "%"+options.PageInfo.Key+"%")
			}
		}
		query = query.Where(likes)
	}

	// 高级查询
	if options.Where != nil {
		query = query.Where(options.Where)
	}

	// 统计数量
	var _c int64
	query.Count(&_c)
	count = int(_c)

	// 排序
	if options.PageInfo.Order != "" {
		query = query.Order(options.PageInfo.Order)
	} else {
		if options.DefaultOrder != "" {
			query = query.Order(options.DefaultOrder)
		} else {
			// 留空就按时间倒序
			query = query.Order("created_at desc")
		}
	}

	// 分页
	limit := options.PageInfo.GetLimit()
	offset := options.PageInfo.GetOffset()
	err = query.Limit(limit).Offset(offset).Find(&list).Error // test

	return
}
