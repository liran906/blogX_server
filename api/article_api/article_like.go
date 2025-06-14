// Path: ./api/article_api/article_like.go

package article_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"blogX_server/service/message_service"
	"blogX_server/service/redis_service/redis_article"
	"blogX_server/utils/jwts"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (ArticleApi) ArticleLikeView(c *gin.Context) {
	req := c.MustGet("bindReq").(models.IDRequest)
	if req.ID == 0 {
		res.FailWithMsg("未指定文章 ID", c)
		return
	}

	var a models.ArticleModel
	err := global.DB.Take(&a, "id = ? AND status = ?", req.ID, 3).Error
	if err != nil {
		res.Fail(err, "文章不存在", c)
		return
	}

	uid := jwts.MustGetClaimsFromRequest(c).UserID

	// 日志
	log := log_service.GetActionLog(c)
	log.ShowRequest()
	log.ShowResponse()
	log.SetLevel(enum.LogTraceLevel)
	log.SetTitle(fmt.Sprintf("文章点赞+ %d", req.ID))

	var al models.ArticleLikesModel
	err = global.DB.Take(&al, "article_id = ? and user_id = ?", a.ID, uid).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建新的点赞记录
			al = models.ArticleLikesModel{
				ArticleID: a.ID,
				UserID:    uid,
			}
			err = global.DB.Create(&al).Error
			if err != nil {
				res.Fail(err, "点赞失败", c)
				return
			}
			// redis文章点赞数+1
			redis_article.AddArticleLike(req.ID)
			res.SuccessWithMsg("点赞成功", c)

			// 通知点赞
			al.ArticleModel = a
			err = message_service.SendArticleLikeNotify(al)
			if err != nil {
				log.SetItemWarn("消息发送失败", err.Error())
			}
			return
		}
		res.Fail(err, "读取点赞数据失败", c)
		return
	}
	err = global.DB.Delete(&al).Error
	if err != nil {
		res.Fail(err, "取消点赞失败", c)
		return
	}
	// redis文章点赞数-1
	redis_article.SubArticleLike(req.ID)
	log.SetTitle(fmt.Sprintf("文章点赞- %d", req.ID))
	res.SuccessWithMsg("取消点赞成功", c)
	return
}
