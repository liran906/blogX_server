// Path: ./api/user_api/user_brief_info.go

package user_api

import (
	"blogX_server/common/res"
	"blogX_server/core"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum/relationship_enum"
	"blogX_server/service/focus_service"
	"blogX_server/service/redis_service/redis_user"
	"blogX_server/utils/jwts"
	"github.com/gin-gonic/gin"
)

type UserBriefInfoResponse struct {
	UserID             uint                       `json:"userID"`
	Nickname           string                     `json:"nickname"`
	AvatarURL          string                     `json:"avatarURL"`
	IPLocation         string                     `json:"ipLocation"`
	ArticleCount       int                        `json:"articleCount"`
	ReadCount          int                        `json:"readCount"`
	LikeCount          int                        `json:"likeCount"`
	CollectCount       int                        `json:"collectCount"`
	FansCount          int                        `json:"fansCount"`
	FollowingCount     int                        `json:"followingCount"`
	SiteAge            int                        `json:"siteAge"`
	ThemeID            uint8                      `json:"themeID"`            // 主页样式 id
	DisplayCollections bool                       `json:"displayCollections"` // 公开我的收藏
	DisplayFans        bool                       `json:"displayFans"`        // 公开我的粉丝
	DisplayFollowing   bool                       `json:"displayFollowing"`   // 公开我的关注
	HomePageVisitCount int                        `json:"homePageVisitCount"` // 主页访问量
	Relation           relationship_enum.Relation `json:"relation"`           // 与登录人的关系
}

func (UserApi) UserBriefInfoView(c *gin.Context) {
	req := c.MustGet("bindReq").(models.IDRequest)

	var u models.UserModel
	err := global.DB.Preload("UserConfigModel").Preload("ArticleModels").Take(&u, req.ID).Error
	if err != nil {
		res.FailWithMsg("用户不存在: "+err.Error(), c)
		return
	}
	redis_user.UpdateHPVCount(u.UserConfigModel)

	loc, _ := core.GetLocationFromIP(u.LastLoginIP)
	var resp = UserBriefInfoResponse{
		UserID:             u.ID,
		Nickname:           u.Nickname,
		AvatarURL:          u.AvatarURL,
		IPLocation:         loc,
		ArticleCount:       len(u.ArticleModels),
		SiteAge:            u.SiteAge(),
		ThemeID:            u.UserConfigModel.ThemeID,
		DisplayCollections: u.UserConfigModel.DisplayCollections,
		DisplayFans:        u.UserConfigModel.DisplayFans,
		DisplayFollowing:   u.UserConfigModel.DisplayFollowing,
		HomePageVisitCount: u.UserConfigModel.HomepageVisitCount + redis_user.GetHPVCount(u.ID),
		FansCount:          0,
		FollowingCount:     0,
	}

	claims, err := jwts.ParseTokenFromRequest(c)
	if err == nil && claims != nil {
		resp.Relation = focus_service.CalcUserRelationship(claims.UserID, req.ID)
	}

	var focusList []models.UserFocusModel
	global.DB.Find(&focusList, "user_id = ? or focus_user_id = ?", req.ID, req.ID)
	for _, model := range focusList {
		if model.UserID == req.ID {
			resp.FollowingCount++
		}
		if model.FocusUserID == req.ID {
			resp.FansCount++
		}
	}

	redis_user.IncreaseHPVCount(req.ID) // 增加主页访问量
	res.SuccessWithData(resp, c)
}
