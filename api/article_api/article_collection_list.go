// Path: ./api/article_api/article_collection_list.go

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
)

type ArticleCollectionListReq struct {
	common.PageInfo
	CollectionID uint               `form:"collectionID" binding:"required"`
	Status       enum.ArticleStatus `form:"status"`
}

type ArticleCollectionListResp struct {
	models.ArticleModel
	UserNickname  string `json:"userNickname,omitempty"`
	UserAvatarURL string `json:"userAvatarURL,omitempty"`
}

// ArticleCollectionListView 某个收藏夹的文章列表
func (ArticleApi) ArticleCollectionListView(c *gin.Context) {
	req := c.MustGet("bindReq").(ArticleCollectionListReq)

	var cf models.CollectionFolderModel
	err := global.DB.Preload("UserConfigModel").Take(&cf, req.CollectionID).Error
	if err != nil {
		res.Fail(err, "收藏夹不存在", c)
		return
	}

	// 提取身份信息，判断查询种类
	claims, err := jwts.ParseTokenFromRequest(c)
	if err != nil || claims == nil {
		// 对未登录搜索限制
		req.PageInfo.Order = ""
		req.StartTime = ""
		req.EndTime = ""
		if req.PageInfo.Page > 1 || req.PageInfo.Limit > 10 {
			res.FailWithMsg("登录后查看更多", c)
			return
		}
	}

	// 校验是否公开收藏夹
	fmt.Println(cf.UserConfigModel)
	if (claims == nil || claims.Role != enum.AdminRoleType) && !cf.UserConfigModel.DisplayCollections {
		res.FailWithMsg("对方未公开收藏夹", c)
		return
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
	}

	// 解析时间戳并查询
	query, err := common.TimeQuery(req.StartTime, req.EndTime)
	if err != nil {
		res.FailWithMsg(err.Error(), c)
		return
	}

	// 获取文章 id 列表
	var articleIDList []uint
	err = global.DB.Model(&models.ArticleCollectionModel{}).
		Where("collection_folder_id = ?", req.CollectionID).
		Select("article_id").Scan(&articleIDList).Error
	if err != nil {
		res.Fail(err, "查询数据库失败", c)
		return
	}
	// 文章 id 列表加入查询
	query = query.Where("id IN ?", articleIDList)

	req.PageInfo.Normalize()

	// 发布文章查询
	_list, count, err := common.ListQuery(
		models.ArticleModel{},
		common.Options{
			PageInfo: req.PageInfo,
			Likes:    []string{"title"},
			Preloads: []string{"UserModel", "CategoryModel"},
			Where:    query,
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
		data := ArticleListResp{                                 // 响应结构体
			ArticleModel:  article,
			UserNickname:  article.UserModel.Nickname,
			UserAvatarURL: article.UserModel.AvatarURL,
		}
		// 非管理员，如果不是自己的文章，看不到状态不为 3 的文章
		// 也就是说 自己的文章，哪怕在别人的收藏夹中，状态为 1234 时候自己都可以看到
		if article.Status != enum.ArticleStatusPublish {
			if claims == nil || (claims.Role != enum.AdminRoleType && article.UserID != claims.UserID) {
				count--
				continue
			}
		}
		list = append(list, data)
	}
	res.SuccessWithList(list, count, c)
}
