// Path: ./api/notify_api/notify_read.go

package notify_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/models/enum/notify_enum"
	"blogX_server/service/log_service"
	"blogX_server/utils/jwts"
	"fmt"
	"github.com/gin-gonic/gin"
)

type NotifyReadReq struct {
	NotifyID   uint `json:"id"` // 读一篇（留空则代表是批量读取）
	NotifyType int8 `json:"t"`  // 批量读取特定类型的消息：1-评论与回复 2-赞和收藏 3-系统通知
}

// NotifyReadView 将消息设为已读
func (NotifyApi) NotifyReadView(c *gin.Context) {
	req := c.MustGet("bindReq").(NotifyReadReq)
	claims := jwts.MustGetClaimsFromRequest(c)

	query := global.DB.Debug().Model(&models.NotifyModel{}).Where("receive_user_id = ? AND is_read = ?", claims.UserID, false)
	if req.NotifyID != 0 {
		// 只要传入了 id，就按照读一篇操作
		query = query.Where("id = ?", req.NotifyID)
		// 不需要验证是否是自己的消息以及是否已读了，因为上面已经在 where 中作为条件被限定了
		// 如果不和规矩，结果就是搜不出来
	} else if req.NotifyType != 0 {
		// 一键已读
		switch req.NotifyType {
		case 1: // 评论与回复
			query = query.Where("type = ? OR type = ?", notify_enum.ArticleCommentType, notify_enum.CommentReplyType)
		case 2: // 赞和收藏
			query = query.Where("type = ? OR type = ? OR type = ?", notify_enum.ArticleLikeType, notify_enum.ArticleCollectType, notify_enum.CommentLikeType)
		case 3: // 系统通知
			query = query.Where("type = ?", notify_enum.SystemType)
		default:
			res.FailWithMsg("type 必须是 1 or 2 or 3", c)
			return
		}
	}

	tx := query.Update("is_read", true)
	if tx.Error != nil {
		res.Fail(tx.Error, "已读失败", c)
		return
	}
	if tx.RowsAffected == 0 {
		res.FailWithMsg("没有符合条件的消息", c)
		return
	}

	log := log_service.GetActionLog(c)
	log.ShowRequest()
	log.ShowResponse()
	log.SetLevel(enum.LogTraceLevel)
	log.SetTitle(fmt.Sprintf("读取消息%d条", tx.RowsAffected))

	res.SuccessWithMsg(fmt.Sprintf("已读 %d 条", tx.RowsAffected), c)
}
