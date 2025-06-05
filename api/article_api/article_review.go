// Path: ./api/article_api/article_review.go

package article_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
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
		res.Fail(err, "审核失败", c)
		return
	}

	// 日志
	log := log_service.GetActionLog(c)
	log.ShowAll()

	// TODO 给文章提交的用户发送消息
	if req.Msg == "" && req.Status == enum.ArticleStatusPublish {
		req.Msg = fmt.Sprintf("您发布审核的文章 [ID:%d]%s 已成功通过审核！", a.ID, a.Title)
		log.SetTitle("文章审核成功")
	}
	if req.Msg == "" && req.Status == enum.ArticleStatusPublish {
		req.Msg = fmt.Sprintf("您发布审核的文章 [ID:%d]%s 没有通过审核！", a.ID, a.Title)
		log.SetTitle("文章审核失败")
	}

	res.SuccessWithMsg("成功审核", c)
}
