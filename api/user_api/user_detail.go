// Path: ./api/user_api/user_detail.go

package user_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/utils/jwts"
	"github.com/gin-gonic/gin"
	"time"
)

type UserDetailResponse struct {
	ID                 uint                    `json:"id"`
	CreatedAt          time.Time               `json:"createdAt"`
	Username           string                  `json:"username"`
	Email              string                  `json:"email"`
	Nickname           string                  `json:"nickname"`
	AvatarURL          string                  `json:"avatarURL"`
	Bio                string                  `json:"bio"`
	OpenID             string                  `json:"openid"`
	Gender             int8                    `json:"gender"`
	Phone              string                  `json:"phone"`
	Country            string                  `json:"country"`
	Province           string                  `json:"province"`
	City               string                  `json:"city"`
	Status             int8                    `json:"status"`
	LastLoginTime      time.Time               `json:"lastLoginTime"`
	LastLoginIP        string                  `json:"lastLoginIP"`
	RegisterSource     enum.RegisterSourceType `json:"registerSource"`
	DateOfBirth        time.Time               `json:"dateOfBirth"`
	Role               enum.RoleType           `json:"role"`               // 角色 1管理员 2普通用户 3访客
	SiteAge            uint                    `json:"siteAge"`            // 站龄
	Tags               []string                `json:"tags"`               // 兴趣标签
	UpdatedAt          *time.Time              `json:"updatedAt"`          // 上次修改时间，可能为空，所以是指针
	ThemeID            uint8                   `json:"themeID"`            // 主页样式 id
	DisplayCollections bool                    `json:"displayCollections"` // 公开我的收藏
	DisplayFans        bool                    `json:"displayFans"`        // 公开我的粉丝
	DisplayFollowing   bool                    `json:"displayFollowing"`   // 公开我的关注
}
type OtherUserDetailResponse struct {
	ID            uint          `json:"id"`
	CreatedAt     time.Time     `json:"createdAt"`
	Username      string        `json:"username"`
	Nickname      string        `json:"nickname"`
	AvatarURL     string        `json:"avatarURL"`
	Bio           string        `json:"bio"`
	Gender        int8          `json:"gender"`
	Country       string        `json:"country"`
	Province      string        `json:"province"`
	City          string        `json:"city"`
	Status        int8          `json:"status"`
	LastLoginTime time.Time     `json:"lastLoginTime"`
	SiteAge       uint          `json:"siteAge"`   // 站龄
	Role          enum.RoleType `json:"role"`      // 角色 1管理员 2普通用户 3访客
	Tags          []string      `json:"tags"`      // 兴趣标签
	UpdatedAt     *time.Time    `json:"updatedAt"` // 上次修改时间，可能为空，所以是指针
	ThemeID       uint8         `json:"themeID"`   // 主页样式 id
}

func (UserApi) UserDetailView(c *gin.Context) {
	req := c.MustGet("bindReq").(models.IDRequest)

	claims, ok := jwts.GetClaimsFromGin(c)
	if !ok {
		res.FailWithMsg("获取用户信息错误，请重新登录", c)
		return
	}
	uid := claims.UserID
	role := claims.Role

	// 读库
	var u models.UserModel
	err := global.DB.Preload("UserConfigModel").Take(&u, "id = ?", req.ID).Error
	if err != nil {
		res.FailWithMsg("读取用户信息失败: "+err.Error(), c)
		return
	}

	// 如果是自己看自己，或者是管理员看任何人，都能看到完整信息
	if req.ID == uid || role == enum.AdminRoleType {
		var resp = UserDetailResponse{
			ID:             u.ID,
			CreatedAt:      u.CreatedAt,
			Username:       u.Username,
			Email:          u.Email,
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
			Role:           u.Role,
			SiteAge:        u.SiteAge(),
		}
		// 判断空指针的情况
		if u.UserConfigModel != nil || u.UserConfigID != 0 {
			resp.Tags = u.UserConfigModel.Tags
			resp.UpdatedAt = u.UserConfigModel.UpdatedAt
			resp.ThemeID = u.UserConfigModel.ThemeID
			resp.DisplayCollections = u.UserConfigModel.DisplayCollections
			resp.DisplayFans = u.UserConfigModel.DisplayFans
			resp.DisplayFollowing = u.UserConfigModel.DisplayFollowing
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
			LastLoginTime: u.LastLoginTime,
			Role:          u.Role,
			SiteAge:       u.SiteAge(),
		}
		// 判断空指针的情况
		if u.UserConfigModel != nil || u.UserConfigID != 0 {
			resp.Tags = u.UserConfigModel.Tags
			resp.UpdatedAt = u.UserConfigModel.UpdatedAt
			resp.ThemeID = u.UserConfigModel.ThemeID
		}
		res.Success(resp, "读取成功", c)
	}
}
