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
	UserID                 uint `json:"userID"`
	ReceiveCommentNotify   bool `json:"receiveCommentNotify"`
	ReceiveLikeNotify      bool `json:"receiveLikeNotify"`
	ReceiveCollectNotify   bool `json:"receiveCollectNotify"`
	ReceivePrivateMessage  bool `json:"receivePrivateMessage"`
	ReceiveStrangerMessage bool `json:"receiveStrangerMessage"`
}

func (NotifyApi) UserNotifyConfUpdateView(c *gin.Context) {
	req := c.MustGet("bindReq").(UserNotifyConfUpdateReq)
	claims := jwts.MustGetClaimsFromRequest(c)

	if claims.UserID != req.UserID {
		res.FailWithMsg("只能更新自己的配置信息", c)
		return
	}

	var un models.UserMessageConfModel
	err := global.DB.Take(&un, "user_id = ?", claims.UserID).Error
	if err != nil {
		res.Fail(err, "用户配置不存在", c)
		return
	}
}
