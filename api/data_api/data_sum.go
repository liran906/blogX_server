// Path: ./api/data_api/data_sum.go

package data_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/redis_service/redis_site"
	"github.com/gin-gonic/gin"
	"time"
)

type SummaryResponse struct {
	FlowCount     int   `json:"flowCount"`
	ClickCount    int   `json:"clickCount"`
	UserCount     int64 `json:"userCount"`
	ArticleCount  int64 `json:"articleCount"`
	MessageCount  int64 `json:"messageCount"`
	CommentCount  int64 `json:"commentCount"`
	NewLoginCount int64 `json:"newLoginCount"`
	NewSignCount  int64 `json:"newSignCount"`
}

func (DataApi) SiteSummaryView(c *gin.Context) {
	var data SummaryResponse

	global.DB.Model(models.UserModel{}).Count(&data.UserCount)
	global.DB.Model(models.ArticleModel{}).Where("status = ?", enum.ArticleStatusPublish).Count(&data.ArticleCount)
	global.DB.Model(models.CommentModel{}).Count(&data.CommentCount)
	global.DB.Model(models.LogModel{}).Where("log_type = ?", enum.LoginLogType).Where("date(created_at) = date(now())").Count(&data.NewLoginCount)
	global.DB.Model(models.UserModel{}).Where("date(created_at) = date(now())").Count(&data.NewSignCount)

	data.FlowCount = redis_site.GetFlow(time.Now().Format("2006-01-02"))
	data.ClickCount = redis_site.GetClick(time.Now().Format("2006-01-02"))

	res.SuccessWithData(data, c)
}
