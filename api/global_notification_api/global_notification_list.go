// Path: ./api/global_notification_api/global_notification_list.go

package global_notification_api

import (
	"blogX_server/common"
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/utils/jwts"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GNListReq struct {
	common.PageInfo
}

type GNListForUserResp struct {
	models.GlobalNotificationModel
	IsRead bool `json:"isRead"`
}

func (GlobalNotificationApi) GNListView(c *gin.Context) {
	req := c.MustGet("bindReq").(GNListReq)
	claims := jwts.MustGetClaimsFromRequest(c)

	// 只有用户侧会用到
	query := global.DB.Where("")
	isReadMap := map[uint]struct{}{}

	// 用户侧查询: 用户已删除的不展示 用户是否已读也要展示出来
	if claims.Role == enum.UserRoleType {
		// 首先把用户全局表的信息读取出来
		var ugnList []models.UserGlobalNotificationModel
		err := global.DB.Find(&ugnList, "user_id = ?", claims.UserID).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			res.Fail(err, "数据库查询失败", c)
			return
		}
		// 处理数据
		var delIDList []uint
		for _, ugn := range ugnList {
			// 出现在用户表里就是已读
			isReadMap[ugn.GlobalNotificationID] = struct{}{}
			// 把用户删除的标记出来
			if ugn.IsDeleted {
				delIDList = append(delIDList, ugn.GlobalNotificationID)
			}
		}
		// 把用户删除的从 query 中排除
		if len(delIDList) > 0 {
			query = query.Where("id NOT IN (?)", delIDList)
		}
	}

	// 解析时间戳并查询
	var err error
	if req.StartTime != "" || req.EndTime != "" {
		query, err = common.TimeQueryWithBase(query, req.StartTime, req.EndTime)
		if err != nil {
			res.FailWithMsg(err.Error(), c)
			return
		}
	}

	// 查询
	_list, count, err := common.ListQuery(models.GlobalNotificationModel{},
		common.Options{
			PageInfo: req.PageInfo,
			Where:    query,
			Likes:    []string{"title", "content"},
			Debug:    false,
		})
	if err != nil {
		res.Fail(err, "查询失败", c)
		return
	}

	// 管理员直接返回就行了
	if claims.Role == enum.AdminRoleType {
		res.SuccessWithList(_list, count, c)
		return
	}

	// 用户这边要加上是否已读
	var list []GNListForUserResp
	for _, gn := range _list {
		_, exists := isReadMap[gn.ID]
		list = append(list, GNListForUserResp{
			GlobalNotificationModel: gn,
			IsRead:                  exists,
		})
	}
	res.SuccessWithList(list, count, c)
}
