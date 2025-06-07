// Path: ./api/comment_api/comment_create.go

package comment_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/service/comment_service"
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

func (commentApi *CommentApi) CommentCreateView(c *gin.Context) {
	req := c.MustGet("bindReq").(CommentCreateReq)
	claims := jwts.MustGetClaimsFromGin(c)

	var a models.ArticleModel
	err := global.DB.Take(&a, req.ArticleID).Error
	if err != nil {
		res.Fail(err, "文章不存在", c)
		return
	}

	req.Content = xss.Filter(req.Content)

	rootID, err := comment_service.GetRootCommentID(req.ParentID)
	if err != nil {
		res.Fail(err, "获取根评论失败", c)
		return
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
