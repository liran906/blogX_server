// Path: ./api/notify_api/user_notify_update.go

package notify_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/utils/jwts"
	"github.com/gin-gonic/gin"
)

// 这个方法在 user config 的相关方法中已经实现了

type UserNotifyConfUpdateReq struct {
	ReceiveCommentNotify   bool `json:"receiveCommentNotify"`
	ReceiveLikeNotify      bool `json:"receiveLikeNotify"`
	ReceiveCollectNotify   bool `json:"receiveCollectNotify"`
	ReceivePrivateMessage  bool `json:"receivePrivateMessage"`
	ReceiveStrangerMessage bool `json:"receiveStrangerMessage"`
}

func (NotifyApi) UserNotifyConfUpdateView(c *gin.Context) {
	req := c.MustGet("bindReq").(UserNotifyConfUpdateReq)
	claims := jwts.MustGetClaimsFromRequest(c)

	var un models.UserMessageConfModel
	err := global.DB.Take(&un, "user_id = ?", claims.UserID).Error
	if err != nil {
		res.Fail(err, "用户配置不存在", c)
		return
	}

	umap := map[string]interface{}{
		"receive_comment_notify":   req.ReceiveCommentNotify,
		"receive_like_notify":      req.ReceiveLikeNotify,
		"receive_collect_notify":   req.ReceiveCollectNotify,
		"receive_private_message":  req.ReceivePrivateMessage,
		"receive_stranger_message": req.ReceiveStrangerMessage,
	}
	err = global.DB.Model(&un).Updates(umap).Error
	if err != nil {
		res.Fail(err, "更新失败", c)
		return
	}
	res.SuccessWithMsg("更新成功", c)
}
