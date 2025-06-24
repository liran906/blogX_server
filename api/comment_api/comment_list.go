// Path: ./api/comment_api/comment_list.go

package comment_api

import (
	"blogX_server/common"
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/models/enum/relationship_enum"
	"blogX_server/service/focus_service"
	"blogX_server/service/redis_service/redis_comment"
	"blogX_server/utils/jwts"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

type CommentListReq struct {
	common.PageInfo
	Type      uint `form:"type" binding:"oneof=1 2 3"` // 1-我的文章收到的评论 2-我发的评论 3-管理员
	ArticleId uint `form:"articleID"`
	UserId    uint `form:"userID"`
}

type CommentListResponse struct {
	ID              uint                       `json:"id"`
	CreatedAt       time.Time                  `json:"createdAt"`
	Content         string                     `json:"content"`
	UserID          uint                       `json:"userID"`
	UserNickname    string                     `json:"userNickname"`
	UserAvatarURL   string                     `json:"userAvatarURL"`
	ArticleID       uint                       `json:"articleID"`
	ArticleTitle    string                     `json:"articleTitle"`
	ArticleCoverURL string                     `json:"articleCoverURL"`
	LikeCount       int                        `json:"likeCount"`
	Relation        relationship_enum.Relation `json:"relation,omitempty"`
	IsMe            bool                       `json:"isMe"`
}

func (CommentApi) CommentListView(c *gin.Context) {
	req := c.MustGet("bindReq").(CommentListReq)
	claims := jwts.MustGetClaimsFromRequest(c)
	// 层级限制
	query := global.DB.Where("depth < ?", global.Config.Site.Article.CommentDepth)

	switch req.Type {
	case 1: // 我的文章收到的评论
		// 查我的文章有哪些
		var alist []uint
		err := global.DB.Model(models.ArticleModel{}).Where("user_id = ?", claims.UserID).Select("id").Scan(&alist).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				res.SuccessWithMsg("您没有发布过文章", c)
				return
			}
			res.Fail(err, "数据库查询失败", c)
			return
		}
		// 高级查询：文章 id 属于我发布过的文章
		query = query.Where("article_id IN ?", alist)
		req.UserId = 0
		req.ArticleId = 0
	case 2: // 我发的评论
		req.UserId = claims.UserID
		req.ArticleId = 0
	case 3: // 管理员查询
		if claims.Role != enum.AdminRoleType {
			res.FailWithMsg("权限不足", c)
			return
		}
		if req.ArticleId != uint(0) {
			err := global.DB.Take(&models.ArticleModel{}, req.ArticleId).Error
			if err != nil {
				res.Fail(err, "文章不存在", c)
				return
			}
		}
		if req.UserId != uint(0) {
			err := global.DB.Take(&models.UserModel{}, req.UserId).Error
			if err != nil {
				res.Fail(err, "用户不存在", c)
				return
			}
		}
	}

	// 解析时间戳并查询
	var err error
	if req.StartTime != "" || req.EndTime != "" {
		query, err = common.TimeQueryWithBase(query, req.StartTime, req.EndTime)
		if err != nil {
			res.FailWithMsg(err.Error(), c)
			return
		}
	}

	req.PageInfo.Normalize()

	_list, count, err := common.ListQuery(models.CommentModel{
		UserID:    req.UserId,
		ArticleID: req.ArticleId,
	}, common.Options{
		PageInfo: req.PageInfo,
		Likes:    []string{"content"},
		Preloads: []string{"UserModel", "ArticleModel"},
		Where:    query,
	})
	if err != nil {
		res.Fail(err, "查询失败", c)
	}

	var relationMap = map[uint]relationship_enum.Relation{}
	if req.Type != 2 {
		var userIDList []uint
		for _, model := range _list {
			userIDList = append(userIDList, model.UserID)
		}
		relationMap = focus_service.CalcUserPatchRelationship(claims.UserID, userIDList)
	}

	var list []CommentListResponse
	for _, cmt := range _list {
		list = append(list, CommentListResponse{
			ID:              cmt.ID,
			CreatedAt:       cmt.CreatedAt,
			Content:         cmt.Content,
			UserID:          cmt.UserID,
			UserNickname:    cmt.UserModel.Nickname,
			UserAvatarURL:   cmt.UserModel.AvatarURL,
			ArticleID:       cmt.ArticleID,
			ArticleTitle:    cmt.ArticleModel.Title,
			ArticleCoverURL: cmt.ArticleModel.CoverURL,
			LikeCount:       cmt.LikeCount + redis_comment.GetCommentLikeCount(cmt.ID),
			Relation:        relationMap[cmt.UserID],
			IsMe:            cmt.UserID == claims.UserID,
		})
	}
	if len(list) == 0 {
		res.SuccessWithMsg("没有相关评论", c)
		return
	}
	res.SuccessWithList(list, count, c)
}
