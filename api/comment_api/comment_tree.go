// Path: ./api/comment_api/comment_tree.go

package comment_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/comment_service"
	"blogX_server/utils/jwts"
	"github.com/gin-gonic/gin"
)

func (CommentApi) CommentTreeView(c *gin.Context) {
	req := c.MustGet("bindReq").(models.IDRequest)

	var article models.ArticleModel
	err := global.DB.Take(&article, req.ID).Error
	if err != nil {
		res.Fail(err, "文章不存在", c)
		return
	}

	// 非管理员只能查看`已发布`的文章
	claims, _ := jwts.ParseTokenFromRequest(c)
	if claims.Role != enum.AdminRoleType && article.Status != enum.ArticleStatusPublish {
		res.FailWithMsg("文章不存在", c)
		return
	}

	// 找到文章所有的根评论
	var rootCmts []models.CommentModel
	err = global.DB.Where("article_id = ? AND root_id IS NULL", req.ID).Find(&rootCmts).Error
	if err != nil {
		res.Fail(err, "数据库查询失败", c)
		return
	}

	if len(rootCmts) == 0 {
		res.SuccessWithMsg("该文章目前没有评论", c)
		return
	}

	var list []comment_service.CommentResponse
	for _, cmt := range rootCmts {
		list = append(list, *comment_service.PreloadAllChildrenResponseFromModel(&cmt))
	}
	res.SuccessWithList(list, len(rootCmts), c)
}
