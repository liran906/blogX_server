// Path: ./api/user_api/user_detail.go

package user_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/redis_service/redis_article"
	"blogX_server/service/redis_service/redis_user"
	"blogX_server/utils/jwts"
	"github.com/gin-gonic/gin"
	"time"
)

type UserDetailResponse struct {
	ID                     uint                    `json:"id"`
	CreatedAt              time.Time               `json:"createdAt"`
	Username               string                  `json:"username"`
	Email                  string                  `json:"email"`
	HasPassword            bool                    `json:"hasPassword"`
	Nickname               string                  `json:"nickname"`
	AvatarURL              string                  `json:"avatarURL"`
	Bio                    string                  `json:"bio"`
	OpenID                 string                  `json:"openid"`
	Gender                 int8                    `json:"gender"`
	Phone                  string                  `json:"phone"`
	Country                string                  `json:"country"`
	Province               string                  `json:"province"`
	City                   string                  `json:"city"`
	Status                 int8                    `json:"status"`
	LastLoginTime          time.Time               `json:"lastLoginTime"`
	LastLoginIP            string                  `json:"lastLoginIP"`
	RegisterSource         enum.RegisterSourceType `json:"registerSource"`
	DateOfBirth            time.Time               `json:"dateOfBirth"`
	ArticleCount           int                     `json:"articleCount"`
	ReadCount              int                     `json:"readCount"`
	LikeCount              int                     `json:"likeCount"`
	CollectCount           int                     `json:"collectCount"`
	FansCount              int                     `json:"fansCount"`
	FollowingCount         int                     `json:"followingCount"`
	Role                   enum.RoleType           `json:"role"`               // 角色 1管理员 2普通用户 3访客
	SiteAge                int                     `json:"siteAge"`            // 站龄
	Tags                   []string                `json:"tags"`               // 兴趣标签
	UpdatedAt              *time.Time              `json:"updatedAt"`          // 上次修改时间，可能为空，所以是指针
	ThemeID                uint8                   `json:"themeID"`            // 主页样式 id
	DisplayCollections     bool                    `json:"displayCollections"` // 公开我的收藏
	DisplayFans            bool                    `json:"displayFans"`        // 公开我的粉丝
	DisplayFollowing       bool                    `json:"displayFollowing"`   // 公开我的关注
	ReceiveCommentNotify   bool                    `json:"receiveCommentNotify"`
	ReceiveLikeNotify      bool                    `json:"receiveLikeNotify"`
	ReceiveCollectNotify   bool                    `json:"receiveCollectNotify"`
	ReceivePrivateMessage  bool                    `json:"receivePrivateMessage"`
	ReceiveStrangerMessage bool                    `json:"receiveStrangerMessage"`
	HomepageVisitCount     int                     `json:"homepageVisitCount"`
}
type OtherUserDetailResponse struct {
	ID                 uint       `json:"id"`
	CreatedAt          time.Time  `json:"createdAt"`
	Username           string     `json:"username"`
	Nickname           string     `json:"nickname"`
	AvatarURL          string     `json:"avatarURL"`
	Bio                string     `json:"bio"`
	Gender             int8       `json:"gender"`
	Country            string     `json:"country"`
	Province           string     `json:"province"`
	City               string     `json:"city"`
	Status             int8       `json:"status"`
	LastLoginTime      time.Time  `json:"lastLoginTime"`
	ArticleCount       int        `json:"articleCount"`
	ReadCount          int        `json:"readCount"`
	LikeCount          int        `json:"likeCount"`
	CollectCount       int        `json:"collectCount"`
	FansCount          int        `json:"fansCount"`
	FollowingCount     int        `json:"followingCount"`
	SiteAge            int        `json:"siteAge"`   // 站龄
	Role               string     `json:"role"`      // 角色 1管理员 2普通用户 3访客
	Tags               []string   `json:"tags"`      // 兴趣标签
	UpdatedAt          *time.Time `json:"updatedAt"` // 上次修改时间，可能为空，所以是指针
	ThemeID            uint8      `json:"themeID"`   // 主页样式 id
	HomepageVisitCount int        `json:"homepageVisitCount"`
}

func (UserApi) UserDetailView(c *gin.Context) {
	req := c.MustGet("bindReq").(models.IDRequest)
	claims := jwts.MustGetClaimsFromRequest(c)
	uid := claims.UserID
	role := claims.Role

	// 传入 id 为 0，就请求自己
	if req.ID == 0 {
		req.ID = claims.UserID
	}

	// 读库
	var u models.UserModel
	err := global.DB.Preload("ArticleModels").Preload("UserConfigModel").Preload("UserMessageConfModel").Take(&u, "id = ?", req.ID).Error
	if err != nil {
		res.FailWithMsg("读取用户信息失败: "+err.Error(), c)
		return
	}
	// 更新缓存中文章信息
	var readCount int
	var likeCount int
	var collectCount int
	for _, a := range u.ArticleModels {
		redis_article.UpdateCachedFieldsForArticle(&a)
		readCount += a.ReadCount
		likeCount += a.LikeCount
		collectCount += a.CollectCount
	}
	// 更新缓存中主页访问量信息
	redis_user.UpdateHPVCount(u.UserConfigModel)

	// 如果是自己看自己，或者是管理员看任何人，都能看到完整信息
	if req.ID == uid || role == enum.AdminRoleType {
		var resp = UserDetailResponse{
			ID:             u.ID,
			CreatedAt:      u.CreatedAt,
			Username:       u.Username,
			Email:          u.Email,
			HasPassword:    u.Password != "",
			Nickname:       u.Nickname,
			AvatarURL:      u.AvatarURL,
			Bio:            u.Bio,
			OpenID:         u.OpenID,
			Gender:         u.Gender,
			Phone:          u.Phone,
			Country:        u.Country,
			Province:       u.Province,
			City:           u.City,
			Status:         u.Status,
			LastLoginTime:  u.LastLoginTime,
			LastLoginIP:    u.LastLoginIP,
			RegisterSource: u.RegisterSource,
			DateOfBirth:    u.DateOfBirth,
			ArticleCount:   len(u.ArticleModels),
			ReadCount:      readCount,
			LikeCount:      likeCount,
			CollectCount:   collectCount,
			Role:           u.Role,
			SiteAge:        u.SiteAge(),
		}
		// 判断空指针的情况
		if u.UserConfigModel != nil {
			resp.Tags = u.UserConfigModel.Tags
			resp.UpdatedAt = u.UserConfigModel.UpdatedAt
			resp.ThemeID = u.UserConfigModel.ThemeID
			resp.DisplayCollections = u.UserConfigModel.DisplayCollections
			resp.DisplayFans = u.UserConfigModel.DisplayFans
			resp.DisplayFollowing = u.UserConfigModel.DisplayFollowing
			resp.HomepageVisitCount = u.UserConfigModel.HomepageVisitCount
		}
		if u.UserMessageConfModel != nil {
			resp.ReceiveCommentNotify = u.UserMessageConfModel.ReceiveCommentNotify
			resp.ReceiveLikeNotify = u.UserMessageConfModel.ReceiveLikeNotify
			resp.ReceiveCollectNotify = u.UserMessageConfModel.ReceiveCollectNotify
			resp.ReceivePrivateMessage = u.UserMessageConfModel.ReceivePrivateMessage
			resp.ReceiveStrangerMessage = u.UserMessageConfModel.ReceiveStrangerMessage
		}
		res.Success(resp, "读取成功", c)
	} else {
		var resp = OtherUserDetailResponse{
			ID:            u.ID,
			CreatedAt:     u.CreatedAt,
			Username:      u.Username,
			Nickname:      u.Nickname,
			AvatarURL:     u.AvatarURL,
			Bio:           u.Bio,
			Gender:        u.Gender,
			Country:       u.Country,
			Province:      u.Province,
			City:          u.City,
			Status:        u.Status,
			ArticleCount:  len(u.ArticleModels),
			ReadCount:     readCount,
			LikeCount:     likeCount,
			CollectCount:  collectCount,
			LastLoginTime: u.LastLoginTime,
			Role:          u.Role.String(),
			SiteAge:       u.SiteAge(),
		}
		// 判断空指针的情况
		if u.UserConfigModel != nil {
			resp.Tags = u.UserConfigModel.Tags
			resp.UpdatedAt = u.UserConfigModel.UpdatedAt
			resp.ThemeID = u.UserConfigModel.ThemeID
			resp.HomepageVisitCount = u.UserConfigModel.HomepageVisitCount
		}
		redis_user.IncreaseHPVCount(u.ID) // 增加主页访问量
		res.Success(resp, "读取成功", c)
	}
}
