// Path: ./api/global_notification_api/global_notification_user_action.go

package global_notification_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"blogX_server/utils/jwts"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (GlobalNotificationApi) GNUserReadView(c *gin.Context) {
	req := c.MustGet("bindReq").(models.IDListRequest)
	claims := jwts.MustGetClaimsFromRequest(c)

	if len(req.IDList) == 0 {
		res.FailWithMsg("没有指定对象", c)
		return
	}
	var gnList []models.GlobalNotificationModel
	err := global.DB.Find(&gnList, "id IN ?", req.IDList).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.Fail(err, "消息不存在", c)
			return
		}
		res.Fail(err, "数据库查询失败", c)
		return
	}
	// 校验 idList
	if len(gnList) == 0 || len(gnList) != len(req.IDList) {
		res.FailWithMsg("请求列表错误", c)
		return
	}

	var inserts []models.UserGlobalNotificationModel
	for _, id := range req.IDList {
		inserts = append(inserts, models.UserGlobalNotificationModel{
			UserID:               claims.UserID,
			GlobalNotificationID: id,
		})
	}

	tx := global.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&inserts)
	if tx.Error != nil || tx.RowsAffected == 0 {
		res.FailWithMsg("标记已读失败", c)
		return
	}
	log := log_service.GetActionLog(c)
	log.ShowRequest()
	log.ShowResponse()
	log.SetLevel(enum.LogTraceLevel)
	log.SetTitle("用户已读全局通知")

	res.SuccessWithMsg(fmt.Sprintf("已读: 共计%d条，成功%d条", len(req.IDList), tx.RowsAffected), c)
}

func (GlobalNotificationApi) GNUserDeleteView(c *gin.Context) {
	req := c.MustGet("bindReq").(models.IDListRequest)
	claims := jwts.MustGetClaimsFromRequest(c)

	if len(req.IDList) == 0 {
		res.FailWithMsg("没有指定对象", c)
		return
	}
	var gnList []models.GlobalNotificationModel
	err := global.DB.Find(&gnList, "id IN ?", req.IDList).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.Fail(err, "消息不存在", c)
			return
		}
		res.Fail(err, "数据库查询失败", c)
		return
	}
	// 校验 idList
	if len(gnList) == 0 || len(gnList) != len(req.IDList) {
		res.FailWithMsg("请求列表错误", c)
		return
	}

	// 先补全记录，避免没有记录无法更新
	var inserts []models.UserGlobalNotificationModel
	for _, id := range req.IDList {
		inserts = append(inserts, models.UserGlobalNotificationModel{
			UserID:               claims.UserID,
			GlobalNotificationID: id,
		})
	}

	// 创建忽略冲突
	err = global.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&inserts).Error
	if err != nil {
		res.Fail(err, "标记删除失败", c)
		return
	}

	// 批量更新 is_deleted 字段
	err = global.DB.Model(&models.UserGlobalNotificationModel{}).
		Where("user_id = ? AND global_notification_id IN ?", claims.UserID, req.IDList).
		Update("is_deleted", true).Error
	if err != nil {
		res.Fail(err, "删除失败", c)
		return
	}

	log := log_service.GetActionLog(c)
	log.ShowRequest()
	log.ShowResponse()
	log.SetLevel(enum.LogTraceLevel)
	log.SetTitle("用户删除全局通知")

	res.SuccessWithMsg("删除成功", c)
}
