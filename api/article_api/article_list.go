// Path: ./api/article_api/article_list.go

package article_api

import (
	"blogX_server/common"
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/redis_service/redis_article"
	"blogX_server/utils/jwts"
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
)

type ArticleListReq struct {
	common.PageInfo
	UserID     uint               `form:"userID"`
	CategoryID *uint              `form:"categoryID"`
	Status     enum.ArticleStatus `form:"status"`
	StartTime  string             `form:"startTime"` // format "2006-01-02 15:04:05"
	EndTime    string             `form:"endTime"`
}

type ArticleListResp struct {
	models.ArticleModel
	UserNickname  string  `json:"userNickname,omitempty"`
	UserAvatarURL string  `json:"userAvatarURL,omitempty"`
	CategoryName  *string `json:"categoryName,omitempty"`
}

// ArticleListView 某个用户发表的文章列表
func (ArticleApi) ArticleListView(c *gin.Context) {
	req := c.MustGet("bindReq").(ArticleListReq)

	var queryType int8 = 1 // 1-未登录 2-已登录查别人 3-已登录查自己 4-管理员

	// 提取身份信息，判断查询种类
	claims, err := jwts.ParseTokenFromRequest(c)
	if err == nil && claims != nil {
		// 判断queryTpye
		if claims.Role == enum.AdminRoleType {
			queryType = 4
		} else {
			if req.UserID == claims.UserID {
				queryType = 3
			} else {
				queryType = 2
			}
			// 非管理员不指定就查自己
			if req.UserID == 0 {
				req.UserID = claims.UserID
			}
		}
	} else {
		// 不登录不指定 id
		if req.UserID == 0 {
			res.FailWithMsg("未指定查询 id", c)
			return
		}
	}

	// 管理员不指定 id 就是查所有
	if req.UserID == 0 {
		// 发布文章查询
		_list, count, err := common.ListQuery(
			models.ArticleModel{
				CategoryID: req.CategoryID,
				Status:     req.Status,
			},
			common.Options{
				PageInfo: req.PageInfo,
				Likes:    []string{"title"},
				Preloads: []string{"UserModel", "CategoryModel"},
				Debug:    false,
			})
		if err != nil {
			res.Fail(err, "查询失败", c)
			return
		}

		var list []ArticleListResp
		for _, article := range _list {
			article.Content = ""                                     // 正文在 list 中不返回
			_ = redis_article.UpdateCachedFieldsForArticle(&article) // 读取缓存中的数据
			data := ArticleListResp{ // 响应结构体
				ArticleModel:  article,
				UserNickname:  article.UserModel.Nickname,
				UserAvatarURL: article.UserModel.AvatarURL,
			}
			if article.CategoryModel != nil {
				// 如果分类不为空，赋值给响应结构体
				data.CategoryName = &article.CategoryModel.Name
			}
			list = append(list, data)
		}
		res.SuccessWithList(list, count, c)
		return
	}

	var u models.UserModel
	err = global.DB.Take(&u, req.UserID).Error
	if err != nil {
		res.FailWithMsg("用户不存在", c)
		return
	}

	// 搜索限制
	switch queryType {
	case 1: // 未登录
		req.Status = enum.ArticleStatusPublish
		req.PageInfo.Order = ""
		req.StartTime = ""
		req.EndTime = ""
		if req.PageInfo.Page > 1 || req.PageInfo.Limit > 10 {
			res.FailWithMsg("登录后查看更多", c)
			return
		}
	case 2: // 已登录查别人
		req.Status = enum.ArticleStatusPublish
		// 情况 3 和 4 都是最高权限，不做限制
	}

	// 提取出置顶的文章, 其余按日期排序
	var defaultOrder string
	var pinnedArticles []models.UserPinnedArticleModel
	err = global.DB.Where("user_id = ?", u.ID). // 如果想加 .Order(...) 等其他链式操作，就必须把条件提取为 .Where(...) 单独写，否则 .Order(...) 就会被忽略
		Order("`rank` ASC"). // 注意 rank 是 MySQL 的保留关键字，必须用反引号 `rank` 包裹，才能作为字段名使用
		Order("created_at DESC,"). // 换了新方法之后 大家的 rank 都一样了，所以要靠置顶时间判断先后
		Find(&pinnedArticles).Error // 另外，order 要在 find（执行）之前，否则失效
	if err == nil {
		for _, m := range pinnedArticles {
			defaultOrder += fmt.Sprintf("id = %d DESC, ", m.ArticleID)
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
	// 如果用户自己传入了顺序
	if req.Order != "" {
		_, exists := orderColumnMap[strings.ToLower(req.Order)]
		if !exists {
			res.FailWithMsg("不支持的排序方式", c)
			return
		}
		// 置顶还是在最前
		req.Order = fmt.Sprintf("%s%s,", defaultOrder, req.Order)
	}
	// 最后加上时间倒序
	defaultOrder += "created_at DESC"

	// 解析时间戳并查询
	query, err := common.TimeQuery(req.StartTime, req.EndTime)
	if err != nil {
		res.FailWithMsg(err.Error(), c)
		return
	}

	req.PageInfo.Normalize()

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
			Preloads:     []string{"UserModel", "CategoryModel"},
			Where:        query,
			DefaultOrder: defaultOrder,
			Debug:        false,
		})
	if err != nil {
		res.Fail(err, "查询失败", c)
		return
	}

	var list []ArticleListResp
	for _, article := range _list {
		article.Content = ""                                     // 正文在 list 中不返回
		_ = redis_article.UpdateCachedFieldsForArticle(&article) // 读取缓存中的数据
		data := ArticleListResp{ // 响应结构体
			ArticleModel:  article,
			UserNickname:  article.UserModel.Nickname,
			UserAvatarURL: article.UserModel.AvatarURL,
		}
		if article.CategoryModel != nil {
			// 如果分类不为空，赋值给响应结构体
			data.CategoryName = &article.CategoryModel.Name
		}
		list = append(list, data)
	}
	res.SuccessWithList(list, count, c)
}
