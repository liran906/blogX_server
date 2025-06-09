// Path: ./api/article_api/article_review.go

package article_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"blogX_server/service/message_service"
	"fmt"
	"github.com/gin-gonic/gin"
)

type ArticleReviewReq struct {
	ArticleID uint               `json:"articleID" binding:"required"`
	Status    enum.ArticleStatus `json:"status" binding:"oneof=3 4"`
	Msg       string             `json:"msg"`
}

// ArticleReviewView 作为管理员提交审核结果
func (ArticleApi) ArticleReviewView(c *gin.Context) {
	req := c.MustGet("bindReq").(ArticleReviewReq)

	var a models.ArticleModel
	err := global.DB.Take(&a, req.ArticleID).Error
	if err != nil {
		res.Fail(err, "文章不存在", c)
		return
	}

	if a.Status != 2 {
		res.FailWithMsg("该文章不处于待审核状态", c)
		return
	}

	err = global.DB.Model(&a).Update("Status", req.Status).Error
	if err != nil {
		res.Fail(err, "审核提交失败", c)
		return
	}

	// 日志
	log := log_service.GetActionLog(c)
	log.ShowAll()
	log.SetTitle("文章审核")

	if req.Status == enum.ArticleStatusPublish {
		fMsg := fmt.Sprintf("您提交审核的文章 [ID:%d]%s 已成功通过！\n", a.ID, a.Title)
		href := fmt.Sprintf("%s/aritcle/%d", global.Config.System.Addr(), a.ID)
		err = message_service.SendSystemNotify(a.UserID, "文章审核通过", fMsg+req.Msg, a.Title, href)
		if err != nil {
			res.Fail(err, "发送消息失败", c)
			return
		}
	} else if req.Status == enum.AritcleStatusFail {
		fMsg := fmt.Sprintf("您提交审核的文章 [ID:%d]%s 没有通过！\n", a.ID, a.Title)
		href := fmt.Sprintf("%s/aritcle/%d", global.Config.System.Addr(), a.ID)
		err = message_service.SendSystemNotify(a.UserID, "文章审核未通过", fMsg+req.Msg, a.Title, href)
		if err != nil {
			res.Fail(err, "发送消息失败", c)
			return
		}
	}
	res.SuccessWithMsg("审核提交成功", c)
}
