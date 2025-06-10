// Path: ./api/global_notification_api/global_notification_create.go

package global_notification_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/service/log_service"
	"github.com/gin-gonic/gin"
)

type GNCreateReq struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	IconURL string `json:"iconURL"`
	Href    string `json:"href"`
}

// GNCreateView 管理员创建管理消息
func (GlobalNotificationApi) GNCreateView(c *gin.Context) {
	//只有 admin 才能进来
	req := c.MustGet("bindReq").(GNCreateReq)

	err := global.DB.Create(&models.GlobalNotificationModel{
		Title:   req.Title,
		Content: req.Content,
		IconURL: req.IconURL,
		Href:    req.Href,
	}).Error
	if err != nil {
		res.Fail(err, "全局消息创建失败", c)
		return
	}

	log := log_service.GetActionLog(c)
	log.ShowAll()
	log.SetTitle("创建全局消息")

	res.SuccessWithMsg("全局消息创建成功", c)
}
