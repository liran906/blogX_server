// Path: ./api/comment_api/comment_tree.go

package comment_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/models/enum/relationship_enum"
	"blogX_server/service/comment_service"
	"blogX_server/service/focus_service"
	"blogX_server/utils"
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

	var userRelationMap = map[uint]relationship_enum.Relation{}
	var userCommentLikeMap = map[uint]struct{}{}
	claims, err := jwts.ParseTokenFromRequest(c)
	if err == nil && claims != nil {
		// 登录了
		var commentList []models.CommentModel // 文章的评论id列表
		global.DB.Find(&commentList, "article_id = ?", req.ID)

		if len(commentList) > 0 {
			// 查我点赞的评论id列表
			var commentIDList []uint
			var userIDList []uint
			for _, model := range commentList {
				commentIDList = append(commentIDList, model.ID)
				userIDList = append(userIDList, model.UserID)
			}
			userIDList = utils.Unique(userIDList) // 对用户id去重
			userRelationMap = focus_service.CalcUserPatchRelationship(claims.UserID, userIDList)

			var commentDiggList []models.CommentLikesModel
			global.DB.Find(&commentDiggList, "user_id = ? and comment_id in ?", claims.UserID, commentIDList)
			for _, model := range commentDiggList {
				userCommentLikeMap[model.CommentID] = struct{}{}
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
		list = append(list, *comment_service.PreloadAllChildrenResponseFromModel(&cmt, userRelationMap, userCommentLikeMap))
	}
	res.SuccessWithList(list, len(rootCmts), c)
}
