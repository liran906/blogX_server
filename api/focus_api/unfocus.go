// Path: ./api/focus_api/unfocus.go

package focus_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/utils/jwts"
	"github.com/gin-gonic/gin"
)

// UnFocusUserView 登录人取关用户
func (FocusApi) UnFocusUserView(c *gin.Context) {
	req := c.MustGet("bindReq").(FocusUserRequest)

	claims := jwts.MustGetClaimsFromRequest(c)
	if req.FocusUserID == claims.UserID {
		res.FailWithMsg("你无法取关自己", c)
		return
	}
	// 查关注的用户是否存在
	var user models.UserModel
	err := global.DB.Take(&user, req.FocusUserID).Error
	if err != nil {
		res.FailWithMsg("取关用户不存在", c)
		return
	}

	// 查之前是否已经关注过他了
	var focus models.UserFocusModel
	err = global.DB.Take(&focus, "user_id = ? and focus_user_id = ?", claims.UserID, user.ID).Error
	if err != nil {
		res.FailWithMsg("未关注此用户", c)
		return
	}
	// 每天的取关也要有个限度？
	// 取关
	global.DB.Delete(&focus)
	res.SuccessWithMsg("取消关注成功", c)
	return
}
