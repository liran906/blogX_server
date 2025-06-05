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
)

// pinned 字段 0 就是没有被置顶，其他数字就是置顶顺序，1为最顶

type ArticlePinReq struct {
	IDList []uint `json:"idList"` // 这里前端注意，顺序就是排序。如果已有 2，增加 1，那么 3 个都要发过来
}

func (ArticleApi) ArticlePinView(c *gin.Context) {
	req := c.MustGet("bindReq").(ArticlePinReq)

	// 置顶数量限制
	if len(req.IDList) > global.Config.Site.Article.MaxPin {
		res.FailWithMsg("已达到普通用户最大置顶数", c)
		return
	}
	//if len(req.IDList) == 0 {
	//	res.FailWithMsg("没有指定置顶文章", c)
	//	return
	//}

	// 校验请求的文章是否存在
	var aModels []models.ArticleModel
	err := global.DB.Find(&aModels, "id in ?", req.IDList).Error
	if err != nil || len(aModels) != len(req.IDList) {
		res.FailWithMsg("文章不存在", c)
		return
	}

	// 提取用户
	claims, ok := jwts.GetClaimsFromGin(c)
	if !ok {
		res.FailWithMsg("请登录", c)
		return
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
			res.FailWithMsg("没有文章权限", c)
			return
		}
	}

	// 日志
	log := log_service.GetActionLog(c)
	log.ShowRequest()
	log.ShowResponse()
	log.SetLevel(enum.LogTraceLevel)
	log.SetTitle("用户更新置顶")

	// 用事务更新
	err = transaction.UpdateUserPinnedArticles(claims.UserID, req.IDList)
	if err != nil {
		res.FailWithData(err.Error(), "置顶失败", c)
		return
	}
	if len(req.IDList) == 0 {
		res.SuccessWithMsg("取消置顶成功", c)
		return
	}
	res.SuccessWithMsg("置顶成功", c)
}
