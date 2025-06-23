// Path: ./api/focus_api/focus_list.go

package focus_api

import (
	"blogX_server/common"
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum/relationship_enum"
	"blogX_server/service/focus_service"
	"blogX_server/utils/jwts"
	"fmt"
	"github.com/gin-gonic/gin"
)

type FocusUserListRequest struct {
	common.PageInfo
	FocusUserID uint `form:"focusUserID"`
	UserID      uint `form:"userID"` // 查用户的关注
}

func (FocusApi) FocusUserListView(c *gin.Context) {
	req := c.MustGet("bindReq").(FocusUserListRequest)
	claims, err := jwts.ParseTokenFromRequest(c)

	if req.UserID != 0 {
		// 传了用户id，我就查这个人关注的用户列表
		var userConf models.UserConfigModel
		err1 := global.DB.Take(&userConf, "user_id = ?", req.UserID).Error
		if err1 != nil {
			res.FailWithMsg("用户配置信息不存在", c)
			return
		}
		if !userConf.DisplayFollowing {
			res.FailWithMsg("此用户未公开我的关注", c)
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
			query.Where("focus_user_id in ?", userIDList)
		}
	}

	_list, count, _ := common.ListQuery(models.UserFocusModel{
		FocusUserID: req.FocusUserID,
		UserID:      req.UserID,
	}, common.Options{
		PageInfo: req.PageInfo,
		Where:    query,
		Preloads: []string{"FocusUserModel"},
	})

	var m = map[uint]relationship_enum.Relation{}
	if err == nil && claims != nil {
		var userIDList []uint
		for _, i2 := range _list {
			userIDList = append(userIDList, i2.FocusUserID)
		}
		m = focus_service.CalcUserPatchRelationship(claims.UserID, userIDList)

	}

	var list = make([]UserListResponse, 0)
	for _, model := range _list {
		list = append(list, UserListResponse{
			UserID:       model.FocusUserID,
			UserNickname: model.FocusUserModel.Nickname,
			UserAvatar:   model.FocusUserModel.AvatarURL,
			UserAbstract: model.FocusUserModel.Bio,
			Relationship: m[model.FocusUserID],
			CreatedAt:    model.CreatedAt,
		})
	}

	res.SuccessWithList(list, count, c)
}
