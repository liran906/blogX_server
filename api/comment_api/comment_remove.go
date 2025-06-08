// Path: ./api/comment_api/comment_remove.go

package comment_api

import (
	"blogX_server/common/res"
	"blogX_server/common/transaction"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"blogX_server/utils/jwts"
	"fmt"
	"github.com/gin-gonic/gin"
)

func (CommentApi) CommentRemoveView(c *gin.Context) {
	req := c.MustGet("bindReq").(models.IDRequest)
	claims := jwts.MustGetClaimsFromGin(c)

	var cmt models.CommentModel
	err := global.DB.Take(&cmt, req.ID).Error
	if err != nil {
		res.Fail(err, "查询数据库失败", c)
		return
	}

	var article models.ArticleModel
	err = global.DB.Take(&article, cmt.ArticleID).Error
	if err != nil {
		res.Fail(err, "查询文章失败", c)
		return
	}

	// 可以删除评论的三类人：评论的发表者 管理员 文章的所有者
	if cmt.UserID != claims.UserID && claims.Role != enum.AdminRoleType && claims.UserID != article.UserID {
		res.FailWithMsg("权限不足", c)
		return
	}

	log := log_service.GetActionLog(c)
	log.ShowAll()
	log.SetTitle(fmt.Sprintf("删除评论[%d]失败", cmt.ID))
	log.SetItem("评论", fmt.Sprintf("%+v", cmt))

	err = transaction.RemoveComment(&cmt)
	if err != nil {
		res.FailWithError(err, c)
	}

	log.SetTitle(fmt.Sprintf("删除评论[%d]成功", cmt.ID))
	res.SuccessWithMsg("评论删除成功", c)
}
