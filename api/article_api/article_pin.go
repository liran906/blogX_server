// Path: ./api/article_api/article_pin.go

package article_api

import (
	"blogX_server/common/res"
	"blogX_server/common/transaction"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"blogX_server/utils/jwts"
	"github.com/gin-gonic/gin"
	"time"
)

type ArticlePinReq struct {
	IDList []uint `json:"idList"` // 这里前端注意，顺序就是排序。如果已有 2，增加 1，那么 3 个都要发过来
}

func (ArticleApi) ArticlePinView(c *gin.Context) {
	req := c.MustGet("bindReq").(ArticlePinReq)
	claims := jwts.MustGetClaimsFromRequest(c)

	// 置顶数量限制
	if len(req.IDList) > global.Config.Site.Article.MaxPin {
		res.FailWithMsg("已达到普通用户最大置顶数", c)
		return
	}

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

	// 用户权限校验
	for _, a := range aModels {
		// 每次只能修改同一作者的置顶
		var ori = aModels[0].UserID
		if ori != a.UserID {
			res.FailWithMsg("只能修改同一作者的置顶", c)
			return
		}

		// 非管理员只能修改自己的用户置顶
		if a.UserID != claims.UserID && claims.Role != enum.AdminRoleType {
			res.FailWithMsg("没有置顶文章权限", c)
			return
		}
	}

	// 日志
	log := log_service.GetActionLog(c)
	log.ShowRequest()
	log.ShowResponse()
	log.SetLevel(enum.LogTraceLevel)
	log.SetTitle("用户文章置顶")

	// 用事务更新
	err = transaction.UpdateUserPinnedArticlesTx(claims.UserID, req.IDList)
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

type ArticlePinListResp struct {
	Rank            int       `json:"rank"`
	ArticleID       uint      `json:"articleID"`
	CreatedAt       time.Time `json:"createdAt"`
	ArticleTitle    string    `json:"articleTitle"`
	ArticleAbstract string    `json:"articleAbstract"`
}

// ArticlePinListView 获取某个用户的置顶文章
func (ArticleApi) ArticlePinListView(c *gin.Context) {
	uid := c.MustGet("bindReq").(models.IDRequest).ID

	var u models.UserModel
	err := global.DB.Take(&u, uid).Error
	if err != nil {
		res.Fail(err, "用户不存在", c)
		return
	}

	var pinnedArticles []models.UserPinnedArticleModel
	err = global.DB.Preload("ArticleModel").Order("`rank` ASC").Find(&pinnedArticles, "user_id = ?", uid).Error
	if err != nil {
		res.Fail(err, "获取置顶文章失败", c)
		return
	}
	if len(pinnedArticles) == 0 {
		res.SuccessWithMsg("当前用户没有置顶文章", c)
		return
	}
	var list []ArticlePinListResp
	for _, a := range pinnedArticles {
		article := ArticlePinListResp{
			Rank:            a.Rank,
			ArticleID:       a.ArticleID,
			CreatedAt:       a.CreatedAt,
			ArticleTitle:    a.ArticleModel.Title,
			ArticleAbstract: a.ArticleModel.Abstract,
		}
		list = append(list, article)
	}
	res.SuccessWithList(list, len(list), c)
}
