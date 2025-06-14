// Path: ./api/user_api/user_list.go

package user_api

import (
	"blogX_server/common"
	"blogX_server/common/res"
	"blogX_server/core"
	"blogX_server/models"
	"github.com/gin-gonic/gin"
	"time"
)

type UserListReq struct {
	common.PageInfo
	StartTime string `form:"startTime"` // format "2006-01-02 15:04:05"
	EndTime   string `form:"endTime"`
}

type UserListResp struct {
	ID              uint      `json:"id"`
	Username        string    `json:"username"`
	CreatedAt       time.Time `json:"createdAt"`
	Status          int8      `json:"status"`
	Nickname        string    `json:"nickname"`
	AvatarURL       string    `json:"avatarURL"`
	Role            string    `json:"role"`
	ArticleCount    int       `json:"articleCount"`
	SiteAge         int       `json:"siteAge"`
	LastLoginIP     string    `json:"lastLoginIP"`
	LastLoginIPAddr string    `json:"lastLoginIPAddr"`
	LastLoginTime   time.Time `json:"lastLoginTime"`
}

func (UserApi) UserListView(c *gin.Context) {
	req := c.MustGet("bindReq").(UserListReq)
	req.PageInfo.Normalize()

	// 解析时间戳并查询
	query, err := common.TimeQuery(req.StartTime, req.EndTime)
	if err != nil {
		res.FailWithMsg(err.Error(), c)
		return
	}

	_list, count, err := common.ListQuery(models.UserModel{}, common.Options{
		PageInfo: req.PageInfo,
		Preloads: []string{"ArticleModels"},
		Where:    query,
		Debug:    true,
	})
	if err != nil {
		res.Fail(err, "查询失败", c)
		return
	}

	var list []UserListResp
	for _, user := range _list {
		addr, err := core.GetLocationFromIP(user.LastLoginIP)
		if err != nil {
			addr = ""
		}
		item := UserListResp{
			ID:              user.ID,
			Username:        user.Username,
			CreatedAt:       user.CreatedAt,
			Status:          user.Status,
			Nickname:        user.Nickname,
			AvatarURL:       user.AvatarURL,
			Role:            user.Role.String(),
			ArticleCount:    len(user.ArticleModels),
			SiteAge:         user.SiteAge(),
			LastLoginIP:     user.LastLoginIP,
			LastLoginIPAddr: addr,
			LastLoginTime:   user.LastLoginTime,
		}
		list = append(list, item)
	}
	res.SuccessWithList(list, count, c)
}
