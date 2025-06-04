// Path: ./api/article_api/article_list.go

package article_api

import (
	"blogX_server/common"
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/utils/jwts"
	"fmt"
	"github.com/gin-gonic/gin"
)

type ArticleListReq struct {
	common.PageInfo
	UserID          uint               `form:"userID"`
	CategoryID      *uint              `form:"categoryID"`
	Status          enum.ArticleStatus `form:"status"`
	CollectionQuery bool               `form:"viewCollect"` // 查看收藏
	StartTime       string             `form:"startTime"`   // format "2006-01-02 15:04:05"
	EndTime         string             `form:"endTime"`
}

// ArticleListView 某个用户发表（或收藏）的文章列表
func (ArticleApi) ArticleListView(c *gin.Context) {
	req := c.MustGet("bindReq").(ArticleListReq)

	var queryType int8 = 1 // 1-未登录 2-已登录查别人 3-已登录查自己 4-管理员

	// 提取身份信息，判断查询种类
	claims, err := jwts.ParseTokenFromRequest(c)
	if err == nil && claims != nil {
		// 不指定就查自己
		if req.UserID == 0 {
			req.UserID = claims.UserID
		}
		// 判断queryTpye
		if claims.Role == enum.AdminRoleType {
			queryType = 4
		} else if req.UserID == claims.UserID {
			queryType = 3
		} else {
			queryType = 2
		}
	}

	// 不登录不指定 id
	if req.UserID == 0 {
		res.FailWithMsg("未指定查询 id", c)
		return
	}
	var u models.UserModel
	err = global.DB.Preload("UserConfigModel").Take(&u, req.UserID).Error
	if err != nil {
		res.FailWithMsg("用户不存在", c)
		return
	}

	// 使 limit 和 page 合理，省的后面再 return
	req.PageInfo.GetLimit()
	req.PageInfo.GetPage()

	// 搜索限制
	switch queryType {
	case 1: // 未登录
		req.Status = enum.ArticleStatusPublish
		req.CollectionQuery = false
		req.PageInfo.Order = ""
		req.StartTime = ""
		req.EndTime = ""
		if req.PageInfo.Page > 1 || req.PageInfo.Limit > 10 {
			res.FailWithMsg("登录后查看更多", c)
			return
		}
	case 2: // 已登录查别人
		if !u.UserConfigModel.DisplayCollections {
			req.CollectionQuery = false
		}
		req.Status = enum.ArticleStatusPublish
		// 情况 3 和 4 都是最高权限，不做限制
	}

	// 提取出置顶的文章, 其余按日期排序
	var defaultOrder string
	var pinnedArticles []models.UserPinnedArticleModel
	err = global.DB.Where("user_id = ?", u.ID). // 如果想加 .Order(...) 等其他链式操作，就必须把条件提取为 .Where(...) 单独写，否则 .Order(...) 就会被忽略
							Order("`rank` asc").        // 注意 rank 是 MySQL 的保留关键字，必须用反引号 `rank` 包裹，才能作为字段名使用
							Find(&pinnedArticles).Error // 另外，order 要在 find（执行）之前，否则失效
	if err == nil {
		for _, m := range pinnedArticles {
			defaultOrder += fmt.Sprintf("id = %d desc, ", m.ArticleID)
		}
	}
	// 支持的排序方式
	var orderColumnMap = map[string]struct{}{
		"read_count desc":    {},
		"like_count desc":    {},
		"comment_count desc": {},
		"collect_count desc": {},
		"read_count asc":     {},
		"like_count asc":     {},
		"comment_count asc":  {},
		"collect_count asc":  {},
	}
	if req.Order != "" {
		_, ok := orderColumnMap[req.Order]
		if !ok {
			res.FailWithMsg("不支持的排序方式", c)
			return
		}
		// 置顶还是在最前
		req.Order = fmt.Sprintf("%s%s, created_at desc", defaultOrder, req.Order)
	}
	// 最后加上时间倒序
	defaultOrder += "created_at desc"

	// 解析时间戳并查询
	query, err := common.TimeQuery(req.StartTime, req.EndTime)
	if err != nil {
		res.FailWithMsg(err.Error(), c)
		return
	}

	if !req.CollectionQuery {
		// 发布文章查询
		_list, count, err := common.ListQuery(
			models.ArticleModel{
				UserID:     req.UserID,
				CategoryID: req.CategoryID,
				Status:     req.Status,
			},
			common.Options{
				PageInfo:     req.PageInfo,
				Likes:        []string{"title"},
				Where:        query,
				DefaultOrder: defaultOrder,
				Debug:        true,
			})
		if err != nil {
			res.FailWithMsg("查询失败", c)
			return
		}
		res.SuccessWithList(_list, count, c)
	} else {
		// 收藏文章查询
	}
}
