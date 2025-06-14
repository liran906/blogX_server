// Path: ./api/user_api/user_info_update.go

package user_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/service/log_service"
	"blogX_server/utils/jwts"
	"blogX_server/utils/mps"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

type UserInfoUpdateReq struct {
	// 批量更新，需要指针判断是没传还是传空值
	Nickname    *string    `json:"nickname" s-u:"nickname"`
	AvatarURL   *string    `json:"avatarURL" s-u:"avatar_url"`
	Bio         *string    `json:"bio" s-u:"bio"`
	Gender      *int8      `json:"gender" s-u:"gender"`
	Phone       *string    `json:"phone" s-u:"phone"`
	Country     *string    `json:"country" s-u:"country"`
	Province    *string    `json:"province" s-u:"province"`
	City        *string    `json:"city" s-u:"city"`
	DateOfBirth *time.Time `json:"dateOfBirth" s-u:"date_of_birth"`

	Tags               *[]string `json:"tags" s-u-c:"tags"`
	DisplayCollections *bool     `json:"displayCollections" s-u-c:"display_collections"`
	DisplayFans        *bool     `json:"displayFans" s-u-c:"display_fans"`
	DisplayFollowing   *bool     `json:"displayFollowing" s-u-c:"display_following"`
	ThemeID            *uint8    `json:"themeID" s-u-c:"theme_id"`

	ReceiveCommentNotify   *bool `json:"receiveCommentNotify" s-m-c:"receive_comment_notify"`
	ReceiveLikeNotify      *bool `json:"receiveLikeNotify" s-m-c:"receive_like_notify"`
	ReceiveCollectNotify   *bool `json:"receiveCollectNotify" s-m-c:"receive_collect_notify"`
	ReceivePrivateMessage  *bool `json:"receivePrivateMessage" s-m-c:"receive_private_message"`
	ReceiveStrangerMessage *bool `json:"receiveStrangerMessage" s-m-c:"receive_stranger_message"`
}

func (UserApi) UserInfoUpdateView(c *gin.Context) {
	req := c.MustGet("bindReq").(UserInfoUpdateReq)

	claims, ok := jwts.GetClaimsFromRequest(c)
	if !ok {
		res.FailWithMsg("请登录", c)
		return
	}

	// 转为 map 方便更新 db
	userMap := mps.StructToMap(req, "s-u")
	userConfMap := mps.StructToMap(req, "s-u-c")
	userMsgConfMap := mps.StructToMap(req, "s-m-c")

	// 判断是否有更新字段
	if len(userMap) == 0 && len(userConfMap) == 0 && len(userMsgConfMap) == 0 {
		res.FailWithMsg("没有更新字段", c)
		return
	}

	// 更新用户表
	if len(userMap) > 0 {
		var u models.UserModel
		err := global.DB.Take(&u, claims.UserID).Error
		if err != nil {
			res.FailWithError(err, c)
			return
		}

		// 30天只能更新一次nickname
		if req.Nickname != nil {
			// 把时间戳转换为 time
			t := time.Unix(u.NicknameUpdate, 0)
			if time.Now().Sub(t).Hours() < 720 {
				res.FailWithMsg("30天内只允许换一次昵称", c)
				return
			}
			userMap["nickname_update"] = time.Now().Unix()
		}

		// 更新
		err = global.DB.Model(&u).Updates(userMap).Error
		if err != nil {
			res.FailWithMsg("写入数据库失败: "+err.Error(), c)
			return
		}
	}

	fmt.Println(userConfMap)

	// 更新 config 表
	if len(userConfMap) > 0 {
		var uc models.UserConfigModel
		err := global.DB.Take(&uc, claims.UserID).Updates(userConfMap).Error
		if err != nil {
			res.FailWithMsg("写入数据库失败: "+err.Error(), c)
			return
		}
	}

	// 更新 messageConfig 表
	if len(userMsgConfMap) > 0 {
		var umc models.UserMessageConfModel
		err := global.DB.Take(&umc, claims.UserID).Updates(userMsgConfMap).Error
		if err != nil {
			res.FailWithMsg("写入数据库失败: "+err.Error(), c)
			return
		}
	}

	// 日志
	log := log_service.GetActionLog(c)
	log.ShowAll()
	log.SetTitle("更新用户信息")

	res.SuccessWithMsg("更新成功", c)
}
