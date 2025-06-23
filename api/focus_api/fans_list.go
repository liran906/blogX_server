// Path: ./api/focus_api/fans_list.go

package focus_api

import (
	"blogX_server/common"
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/utils/jwts"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

type FansUserListResponse struct {
	FansUserID       uint      `json:"fansUserID"`
	FansUserNickname string    `json:"fansUserNickname"`
	FansUserAvatar   string    `json:"fansUserAvatar"`
	FansUserAbstract string    `json:"fansUserAbstract"`
	CreatedAt        time.Time `json:"createdAt"`
}

// FansUserListView 我的粉丝和用户的粉丝
func (FocusApi) FansUserListView(c *gin.Context) {
	req := c.MustGet("bindReq").(FocusUserListRequest)
	claims, err := jwts.ParseTokenFromRequest(c)

	if req.UserID != 0 {
		// 传了用户id，我就查这个人的粉丝列表
		var userConf models.UserConfigModel
		err1 := global.DB.Take(&userConf, "user_id = ?", req.UserID).Error
		if err1 != nil {
			res.FailWithMsg("用户配置信息不存在", c)
			return
		}
		if !userConf.DisplayFans {
			res.FailWithMsg("此用户未公开我的粉丝", c)
			return
		}
		// 如果你没登录。我就不允许你查第二页
		if err != nil || claims == nil {
			if req.Limit > 10 || req.Page > 1 {
				res.FailWithMsg("未登录用户只能显示第一页", c)
				return
			}
		}
	} else {
		if err != nil || claims == nil {
			res.FailWithMsg("请登录", c)
			return
		}
		req.UserID = claims.UserID
	}

	query := global.DB.Where("")
	if req.Key != "" {
		// 模糊匹配用户
		var userIDList []uint
		global.DB.Model(&models.UserModel{}).
			Where("nickname like ?", fmt.Sprintf("%%%s%%", req.Key)).
			Select("id").Scan(&userIDList)
		if len(userIDList) > 0 {
			query.Where("user_id in ?", userIDList)
		}
	}

	_list, count, _ := common.ListQuery(models.UserFocusModel{
		FocusUserID: req.UserID,
		UserID:      req.FocusUserID,
	}, common.Options{
		PageInfo: req.PageInfo,
		Where:    query,
		Preloads: []string{"UserModel"},
	})

	var list = make([]FansUserListResponse, 0)
	for _, model := range _list {
		list = append(list, FansUserListResponse{
			FansUserID:       model.UserID,
			FansUserNickname: model.UserModel.Nickname,
			FansUserAvatar:   model.UserModel.AvatarURL,
			FansUserAbstract: model.UserModel.Bio,
			CreatedAt:        model.CreatedAt,
		})
	}

	res.SuccessWithList(list, count, c)
}
