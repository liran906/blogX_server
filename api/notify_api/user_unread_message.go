// Path: ./api/notify_api/user_unread_message.go

package notify_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum/notify_enum"
	"blogX_server/utils/jwts"
	"github.com/gin-gonic/gin"
)

type UserUnreadMessageResp struct {
	CommentMsgCount int `json:"commentMsgCount"`
	LikeMsgCount    int `json:"likeMsgCount"`
	PrivateMsgCount int `json:"privateMsgCount"`
	SystemMsgCount  int `json:"systemMsgCount"`
}

// UserUnreadMessageView 查看用户未读的所有消息（站内信、系统通知、私信）数量
func (NotifyApi) UserUnreadMessageView(c *gin.Context) {
	claims := jwts.MustGetClaimsFromGin(c)

	var notifies []models.NotifyModel
	global.DB.Find(&notifies, "receive_user_id = ? AND is_read = ?", claims.UserID, false)

	// 首先读取所有未读通知
	var resp UserUnreadMessageResp
	for _, notify := range notifies {
		switch notify.Type {
		case notify_enum.ArticleCommentType, notify_enum.CommentReplyType:
			resp.CommentMsgCount++
		case notify_enum.ArticleLikeType, notify_enum.ArticleCollectType, notify_enum.CommentLikeType:
			resp.LikeMsgCount++
		case notify_enum.SystemType:
			resp.SystemMsgCount++
		}
	}

	// 然后读取所有未读系统通知
	var total int64
	global.DB.Model(&models.GlobalNotificationModel{}).Count(&total)
	var read int64
	global.DB.Model(&models.UserGlobalNotificationModel{}).Where("user_id = ?", claims.UserID).Count(&read)
	resp.SystemMsgCount += int(total - read)

	// TODO 未读私信数量统计

	res.SuccessWithData(resp, c)
}
