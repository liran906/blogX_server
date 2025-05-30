// Path: ./blogX_server/api/user_api/user_brief_info.go

package user_api

import (
	"blogX_server/common/res"
	"blogX_server/core"
	"blogX_server/global"
	"blogX_server/models"
	"github.com/gin-gonic/gin"
)

type UserBriefInfoResponse struct {
	UserID         uint   `json:"userID"`
	Nickname       string `json:"nickname"`
	AvatarURL      string `json:"avatarURL"`
	IPLocation     string `json:"ipLocation"`
	ViewCount      uint   `json:"viewCount"`
	ArticleCount   uint   `json:"articleCount"`
	FansCount      uint   `json:"fansCount"`
	FollowingCount uint   `json:"followingCount"`
	SiteAge        uint   `json:"siteAge"`
}

func (UserApi) UserBriefInfoView(c *gin.Context) {
	req := c.MustGet("bindReq").(models.IDRequest)

	var u models.UserModel
	err := global.DB.Take(&u, req.ID).Error
	if err != nil {
		res.FailWithMsg("用户不存在: "+err.Error(), c)
		return
	}

	loc, _ := core.GetLocationFromIP(u.LastLoginIP)
	var resp = UserBriefInfoResponse{
		UserID:     u.ID,
		Nickname:   u.Nickname,
		AvatarURL:  u.AvatarURL,
		IPLocation: loc,
		SiteAge:    u.SiteAge(),
		// tbd
		ViewCount:      1,
		ArticleCount:   1,
		FansCount:      1,
		FollowingCount: 1,
	}
	res.SuccessWithData(resp, c)
}
