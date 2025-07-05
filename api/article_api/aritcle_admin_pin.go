// Path: ./api/article_api/aritcle_admin_pin.go

package article_api

import (
	"blogX_server/common/res"
	"blogX_server/common/transaction"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"blogX_server/service/redis_service/redis_cache"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

// ArticleAdminPinView 管理员将文章整体置顶
func (ArticleApi) ArticleAdminPinView(c *gin.Context) {
	// 这里前端注意，ArticlePinReq 顺序就是排序。如果已有 2，增加 1，那么 3 个都要发过来
	req := c.MustGet("bindReq").(ArticlePinReq)

	// 校验请求的文章是否存在
	var aModels []models.ArticleModel
	err := global.DB.Find(&aModels, "id in ?", req.IDList).Error
	if err != nil || len(aModels) != len(req.IDList) {
		res.FailWithMsg("文章不存在", c)
		return
	}
	// 只能置顶已发布的文章
	for _, a := range aModels {
		if a.Status != enum.ArticleStatusPublish {
			res.FailWithMsg("只能置顶已发布的文章", c)
			return
		}
	}

	// 日志
	log := log_service.GetActionLog(c)
	log.ShowRequest()
	log.ShowResponse()
	log.SetLevel(enum.LogTraceLevel)
	log.SetTitle("站点文章置顶")

	// 用事务更新
	err = transaction.UpdateSitePinnedArticlesTx(req.IDList)
	if err != nil {
		res.Fail(err, "置顶失败", c)
		return
	}
	if len(req.IDList) == 0 {
		res.SuccessWithMsg("取消置顶成功", c)
		return
	}
	res.SuccessWithMsg("置顶成功", c)
}

type ArticleAdminPinListResp struct {
	Rank            int       `json:"rank"`
	ArticleID       uint      `json:"articleID"`
	CreatedAt       time.Time `json:"createdAt"`
	ArticleTitle    string    `json:"articleTitle"`
	ArticleAbstract string    `json:"articleAbstract"`
	AuthorID        uint      `json:"authorID"`
}

func (ArticleApi) ArticleAdminPinListView(c *gin.Context) {
	var pinnedArticles []models.UserPinnedArticleModel
	err := global.DB.Preload("ArticleModel").Order("`rank` ASC").Find(&pinnedArticles, "user_id = ?", 0).Error
	if err != nil {
		res.Fail(err, "获取置顶文章失败", c)
		return
	}
	if len(pinnedArticles) == 0 {
		res.SuccessWithMsg("当前没有置顶文章", c)
		return
	}
	var list []ArticleAdminPinListResp
	for _, a := range pinnedArticles {
		article := ArticleAdminPinListResp{
			Rank:            a.Rank,
			ArticleID:       a.ArticleID,
			CreatedAt:       a.CreatedAt,
			ArticleTitle:    a.ArticleModel.Title,
			ArticleAbstract: a.ArticleModel.Abstract,
			AuthorID:        a.ArticleModel.UserID,
		}
		list = append(list, article)
		redis_cache.CacheCloseCertain(fmt.Sprintf("%s%d", redis_cache.CacheArticleDetailPrefix, a.ArticleID))
	}
	res.SuccessWithList(list, len(list), c)
}
