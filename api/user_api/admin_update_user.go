// Path: ./blogX_server/api/user_api/admin_update_user.go

package user_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"blogX_server/utils/mps"
	"github.com/gin-gonic/gin"
	"time"
)

type AdminUpdateUserRequest struct {
	UserID      uint           `json:"userID" binding:"required"`
	Username    *string        `json:"username" s-u:"username"`
	Nickname    *string        `json:"nickname" s-u:"nickname"`
	AvatarURL   *string        `json:"avatarURL" s-u:"avatar_url"`
	Bio         *string        `json:"bio" s-u:"bio"`
	Gender      *int8          `json:"gender" s-u:"gender"`
	Phone       *string        `json:"phone" s-u:"phone"`
	Country     *string        `json:"country" s-u:"country"`
	Province    *string        `json:"province" s-u:"province"`
	City        *string        `json:"city" s-u:"city"`
	DateOfBirth *time.Time     `json:"dateOfBirth" s-u:"date_of_birth"`
	Role        *enum.RoleType `json:"role" s-u:"role"`
}

func (UserApi) AdminUpdateUserView(c *gin.Context) {
	var req AdminUpdateUserRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	userMap := mps.StructToMap(req, "s-u")

	if len(userMap) == 0 {
		res.FailWithMsg("没有更新字段", c)
		return
	}

	err = global.DB.Take(&models.UserModel{}, req.UserID).Updates(userMap).Error
	if err != nil {
		res.FailWithMsg("更新用户信息失败: "+err.Error(), c)
		return
	}

	// 日志
	log := log_service.GetActionLog(c)
	log.ShowAll()
	log.SetTitle("管理员更新用户信息")

	res.SuccessWithMsg("更新用户信息成功", c)
}
