// Path: ./api/comment_api/comment_create.go

package comment_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/comment_service"
	"blogX_server/service/log_service"
	"blogX_server/service/message_service"
	"blogX_server/service/redis_service/redis_article"
	"blogX_server/service/redis_service/redis_comment"
	"blogX_server/utils/jwts"
	"blogX_server/utils/xss"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CommentCreateReq struct {
	Content   string `json:"content" binding:"required"`
	ArticleID uint   `json:"articleID"`
	ParentID  *uint  `json:"parentID"`
}

func (CommentApi) CommentCreateView(c *gin.Context) {
	req := c.MustGet("bindReq").(CommentCreateReq)
	claims := jwts.MustGetClaimsFromRequest(c)

	if req.ArticleID == 0 && req.ParentID == nil {
		res.FailWithMsg("请输入文章 id 或父评论 id", c)
		return
	}

	var err error
	if req.ArticleID != 0 {
		_, err = verifyArticle(req.ArticleID)
		if err != nil {
			res.FailWithError(err, c)
			return
		}
	}

	// 确定深度以及根评论 ID
	var depth = 0
	var rootID *uint
	var parent models.CommentModel
	if req.ParentID != nil {
		err := global.DB.Take(&parent, req.ParentID).Error
		if err != nil {
			res.Fail(err, "获取父评论失败", c)
			return
		}
		// 文章 id
		if req.ArticleID != 0 {
			// 校验文章 id
			if parent.ArticleID != req.ArticleID {
				res.FailWithMsg("文章 id 或父评论 id 错误", c)
				return
			}
		} else {
			// 获取文章 id
			req.ArticleID = parent.ArticleID
		}

		// 本次评论的 root 及 depth
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

	// 校验文章
	article, err := verifyArticle(req.ArticleID)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	req.Content = xss.Filter(req.Content)

	log := log_service.GetActionLog(c)
	log.ShowAll()
	log.SetTitle("创建评论失败")

	// 创建
	var cmt = models.CommentModel{
		UserID:    claims.UserID,
		Content:   req.Content,
		ArticleID: req.ArticleID,
		ParentID:  req.ParentID,
		RootID:    rootID,
		Depth:     depth,
	}

	// 入库
	err = global.DB.Create(&cmt).Error
	if err != nil {
		res.Fail(err, "无法评论该文章", c)
		return
	}

	// 更新祖先评论的回复量
	if cmt.ParentID != nil {
		ancestors, err := comment_service.GetAncestors(*cmt.ParentID)
		if err != nil {
			res.Fail(err, "获取父评论失败", c)
			return
		}
		for _, ans := range ancestors {
			redis_comment.AddCommentReplyCount(ans.ID)
		}
	}

	// 更新文章回复量
	redis_article.AddArticleComment(req.ArticleID)

	log.SetTitle("创建评论成功")
	res.SuccessWithMsg("创建评论成功", c)

	// SendCommentNotify 发送提醒消息
	// ======================================
	// 进去内部函数在 preload  cmt的fk，总是会报错
	// 怀疑是 mysql 没有那么快写入并可读取
	// 所以在外部把相关字段填好再传入吧
	cmt.ParentModel = &parent
	cmt.ArticleModel = article
	err = message_service.SendCommentNotify(cmt)
	if err != nil {
		log.SetItemWarn("消息发送失败", err.Error())
	}
}

func verifyArticle(articleID uint) (article models.ArticleModel, err error) {
	err = global.DB.Take(&article, articleID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("文章不存在")
		}
		return
	}
	// 只能评论开放评论的文章
	if !article.OpenForComment {
		err = errors.New("无法评论该文章")
		return
	}
	// 只能评论已发布的文章
	if article.Status != enum.ArticleStatusPublish {
		err = errors.New("无法评论该文章")
		return
	}
	return
}
