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
	"github.com/sirupsen/logrus"
)

func (CommentApi) CommentTreeView(c *gin.Context) {
	req := c.MustGet("bindReq").(models.IDRequest)

	var article models.ArticleModel
	err := global.DB.Take(&article, req.ID).Error
	if err != nil {
		res.Fail(err, "文章不存在", c)
		return
	}

	var userCommentLikeMap = map[uint]struct{}{}
	claims, err := jwts.ParseTokenFromRequest(c)
	if err == nil && claims != nil {
		var commentLikes []models.CommentLikesModel
		subQuery := global.DB.Model(&models.CommentModel{}).Select("id").Where("article_id = ?", article.ID)
		err = global.DB.Preload("CommentModel").Find(&commentLikes, "user_id = ? AND comment_id IN (?)", claims.UserID, subQuery).Error
		if err != nil {
			logrus.Warnf("查询文章评论数据库失败: %v", err)
		}
		if len(commentLikes) > 0 {
			for _, commentLike := range commentLikes {
				userCommentLikeMap[commentLike.CommentID] = struct{}{}
			}
		}
	}

	// 非管理员只能查看`已发布`的文章
	if (err != nil || claims == nil || claims.Role != enum.AdminRoleType) && article.Status != enum.ArticleStatusPublish {
		res.FailWithMsg("文章不存在", c)
		return
	}

	// 找到文章所有的根评论
	var rootCmts []models.CommentModel
	err = global.DB.Where("article_id = ? AND root_id IS NULL", req.ID).Order("created_at DESC").Find(&rootCmts).Error
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
		list = append(list, *comment_service.PreloadAllChildrenResponseFromModel(&cmt, userCommentLikeMap))
	}
	res.SuccessWithList(list, len(rootCmts), c)
}
