// Path: ./api/comment_api/comment_like.go

package comment_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"blogX_server/service/message_service"
	"blogX_server/service/redis_service/redis_comment"
	"blogX_server/utils/jwts"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (CommentApi) CommentLikeView(c *gin.Context) {
	req := c.MustGet("bindReq").(models.IDRequest)
	if req.ID == 0 {
		res.FailWithMsg("未指定评论 ID", c)
		return
	}

	var cmt models.CommentModel
	err := global.DB.Take(&cmt, req.ID).Error
	if err != nil {
		res.Fail(err, "评论不存在", c)
		return
	}

	uid := jwts.MustGetClaimsFromRequest(c).UserID

	// 日志
	log := log_service.GetActionLog(c)
	log.ShowRequest()
	log.ShowResponse()
	log.SetLevel(enum.LogTraceLevel)
	log.SetTitle(fmt.Sprintf("评论点赞+ %d", req.ID))

	var cl models.CommentLikesModel
	err = global.DB.Take(&cl, "comment_id = ? and user_id = ?", cmt.ID, uid).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			cl = models.CommentLikesModel{
				UserID:    uid,
				CommentID: req.ID,
			}
			err = global.DB.Create(&cl).Error
			if err != nil {
				res.Fail(err, "点赞失败", c)
				return
			}
			// redis 评论点赞数+1
			redis_comment.AddCommentLikeCount(req.ID)
			res.SuccessWithMsg("点赞成功", c)

			// 通知点赞
			cl.CommentModel = cmt
			err = message_service.SendCommentLikeNotify(cl)
			if err != nil {
				log.SetItemWarn("消息发送失败", err.Error())
			}
			return
		}
		res.Fail(err, "读取点赞数据失败", c)
		return
	}
	err = global.DB.Delete(&cl).Error
	if err != nil {
		res.Fail(err, "取消点赞失败", c)
		return
	}
	// redis文章点赞数-1
	redis_comment.SubCommentLikeCount(req.ID)
	log.SetTitle(fmt.Sprintf("评论点赞- %d", req.ID))
	res.SuccessWithMsg("取消点赞成功", c)
	return
}
