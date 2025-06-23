// Path: ./api/focus_api/focus.go

package focus_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/utils/jwts"
	"github.com/gin-gonic/gin"
)

type FocusUserRequest struct {
	FocusUserID uint `json:"focusUserID" binding:"required"`
}

// FocusUserView 当前用户关注其他用户
func (FocusApi) FocusUserView(c *gin.Context) {
	req := c.MustGet("bindReq").(FocusUserRequest)

	claims := jwts.MustGetClaimsFromRequest(c)
	if req.FocusUserID == claims.UserID {
		res.FailWithMsg("你时刻都在关注自己", c)
		return
	}
	// 查关注的用户是否存在
	var user models.UserModel
	err := global.DB.Take(&user, req.FocusUserID).Error
	if err != nil {
		res.FailWithMsg("关注用户不存在", c)
		return
	}

	// 查之前是否已经关注过他了
	var focus models.UserFocusModel
	err = global.DB.Take(&focus, "user_id = ? and focus_user_id = ?", claims.UserID, user.ID).Error
	if err == nil {
		res.FailWithMsg("请勿重复关注", c)
		return
	}

	// 每天关注是不是应该有个限度？
	// 每天的取关也要有个限度？

	// 关注
	global.DB.Create(&models.UserFocusModel{
		UserID:      claims.UserID,
		FocusUserID: req.FocusUserID,
	})

	res.SuccessWithMsg("关注成功", c)
	return
}
