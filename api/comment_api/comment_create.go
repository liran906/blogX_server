// Path: ./api/comment_api/comment_create.go

package comment_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"blogX_server/service/redis_service/redis_article"
	"blogX_server/utils/jwts"
	"blogX_server/utils/xss"
	"github.com/gin-gonic/gin"
)

type CommentCreateReq struct {
	Content   string `json:"content" binding:"required"`
	ArticleID uint   `json:"articleID" binding:"required"`
	ParentID  *uint  `json:"parentID"`
}

func (CommentApi) CommentCreateView(c *gin.Context) {
	req := c.MustGet("bindReq").(CommentCreateReq)
	claims := jwts.MustGetClaimsFromGin(c)

	var article models.ArticleModel
	err := global.DB.Take(&article, req.ArticleID).Error
	if err != nil {
		res.Fail(err, "文章不存在", c)
		return
	}

	// 只能评论已发布的文章
	if article.Status != enum.ArticleStatusPublish {
		res.FailWithMsg("无法评论该文章", c)
		return
	}

	req.Content = xss.Filter(req.Content)

	var depth = 0
	var rootID *uint
	if req.ParentID != nil {
		var parent models.CommentModel
		err := global.DB.Take(&parent, req.ParentID).Error
		if err != nil {
			res.Fail(err, "获取父评论失败", c)
			return
		}
		depth = parent.Depth + 1
		if parent.RootID == nil {
			rootID = &parent.ID
		} else {
			rootID = parent.RootID
		}
		if depth >= global.Config.Site.Article.CommentDepth {
			res.FailWithMsg("评论层级超过限制", c)
			return
		}
	}

	log := log_service.GetActionLog(c)
	log.ShowAll()
	log.SetTitle("创建评论失败")

	var cmt = models.CommentModel{
		UserID:    claims.UserID,
		Content:   req.Content,
		ArticleID: req.ArticleID,
		ParentID:  req.ParentID,
		RootID:    rootID,
		Depth:     depth,
	}

	err = global.DB.Create(&cmt).Error
	if err != nil {
		res.Fail(err, "创建评论失败", c)
		return
	}

	redis_article.AddArticleComment(req.ArticleID)

	log.SetTitle("创建评论成功")
	res.SuccessWithMsg("创建评论成功", c)
}
