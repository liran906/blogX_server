// Path: ./api/article_api/article_remove.go

package article_api

import (
	"blogX_server/common/res"
	"blogX_server/common/transaction"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"blogX_server/utils/jwts"
	"fmt"
	"github.com/gin-gonic/gin"
)

// 管理员随便删，用户只能删自己的
// 如果是物理删除，那需要删除对应记录（点赞、收藏、置顶等）

// ArticleRemoveView 删除单篇
func (ArticleApi) ArticleRemoveView(c *gin.Context) {
	req := c.MustGet("bindReq").(models.IDRequest)
	claims := jwts.MustGetClaimsFromGin(c)

	var a models.ArticleModel
	err := global.DB.Take(&a, req.ID).Error
	if err != nil {
		res.Fail(err, "文章不存在", c)
		return
	}
	if claims.UserID != a.UserID && claims.Role != enum.AdminRoleType {
		res.FailWithMsg("没有此文章的删除权限", c)
		return
	}

	// log
	log := log_service.GetActionLog(c)
	log.ShowAll()
	log.SetLevel(enum.LogWarnLevel)
	log.SetTitle("删除文章")
	if claims.UserID != a.UserID {
		log.ShowClaim(claims)
	}

	mps, err := transaction.RemoveArticleAndRelated(&a)
	if err != nil {
		res.Fail(err, "文章删除失败", c)
		return
	}
	log.SetItem(fmt.Sprintf("删除文章 %d", a.ID), mps)
	res.SuccessWithMsg("文章删除成功", c)
}

// ArticleBatchRemoveView 批量删除，只能管理员
func (ArticleApi) ArticleBatchRemoveView(c *gin.Context) {
	req := c.MustGet("bindReq").(models.RemoveRequest)

	// log
	log := log_service.GetActionLog(c)
	log.ShowAll()
	log.SetLevel(enum.LogWarnLevel)
	log.SetTitle("批量删除文章")

	var succ []uint
	var count int

	for _, aid := range req.IDList {
		var a models.ArticleModel
		err := global.DB.Take(&a, aid).Error
		if err != nil {
			log.SetItemWarn(fmt.Sprintf("删除失败 id:%d", a.ID), "文章不存在: "+err.Error())
			continue
		}
		mps, err := transaction.RemoveArticleAndRelated(&a)
		if err != nil {
			log.SetItemWarn(fmt.Sprintf("删除失败 id:%d", a.ID), "删除事务失败: "+err.Error())
			continue
		}
		log.SetItem(fmt.Sprintf("删除文章 %d", a.ID), mps)
		succ = append(succ, a.ID)
		count++
	}

	var result = map[string]any{
		"total":        len(req.IDList),
		"successCount": count,
		"deleted":      succ,
	}

	if len(succ) == 0 {
		res.FailWithData(result, "文章批量删除失败", c)
		return
	}
	res.Success(result, "文章批量删除成功", c)
}
