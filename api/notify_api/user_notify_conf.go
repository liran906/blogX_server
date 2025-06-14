// Path: ./api/notify_api/user_notify_conf.go

package notify_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/utils/jwts"
	"github.com/gin-gonic/gin"
)

// 这个方法在 user config 的相关方法中已经实现了

func (NotifyApi) UserNotifyConfView(c *gin.Context) {
	claims := jwts.MustGetClaimsFromRequest(c)

	var un models.UserMessageConfModel
	err := global.DB.Take(&un, "user_id = ?", claims.UserID).Error
	if err != nil {
		res.Fail(err, "用户配置不存在", c)
		return
	}

	res.SuccessWithData(un, c)
}
