// Path: ./api/global_notification_api/global_notification_remove.go

package global_notification_api

import (
	"blogX_server/common/res"
	"blogX_server/common/transaction"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/service/log_service"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GNRemoveView 管理员删除全局消息
func (GlobalNotificationApi) GNRemoveView(c *gin.Context) {
	req := c.MustGet("bindReq").(models.IDListRequest)

	if len(req.IDList) == 0 {
		res.FailWithMsg("没有指定删除对象", c)
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
		res.FailWithMsg("删除列表错误", c)
		return
	}

	// 日志
	log := log_service.GetActionLog(c)
	log.ShowAll()
	if len(gnList) > 1 {
		log.SetTitle("批量删除全局通知")
	} else {
		log.SetTitle("删除全局通知")
	}

	var count []uint
	for _, gn := range gnList {
		err := transaction.RemoveGlobalNotificationTx(&gn)
		if err != nil {
			log.SetItemWarn("删除失败", fmt.Sprintf("[%d]删除失败: %s", gn.ID, err.Error()))
			continue
		}
		count = append(count, gn.ID)
	}

	if len(count) == 0 {
		res.FailWithMsg("删除失败", c)
		return
	}
	res.SuccessWithMsg(fmt.Sprintf("全局消息删除 共计%d条 成功%d条 成功删除id:%v", len(count), len(gnList), count), c)
}
