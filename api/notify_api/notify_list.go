// Path: ./api/notify_api/notify_list.go

package notify_api

import (
	"blogX_server/common"
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum/notify_enum"
	"blogX_server/models/enum/relationship_enum"
	"blogX_server/service/focus_service"
	"blogX_server/utils/jwts"
	"github.com/gin-gonic/gin"
	"time"
)

type NotifyListReq struct {
	common.PageInfo `json:"pageInfo"`
	NotifyType      int8 `form:"t" binding:"required,oneof=1 2 3"` // 1-评论与回复 2-赞和收藏 3-系统通知 `
}

type NotifyListResp struct {
	ID                  uint                       `json:"id"`
	CreatedAt           time.Time                  `json:"createdAt"`
	NotifyType          string                     `json:"type"`
	Title               string                     `json:"title"`
	Content             string                     `json:"content,omitempty"`
	RevUserID           uint                       `json:"revUserID"`
	ActionUserID        uint                       `json:"actionUserID,omitempty"`
	ActionUserNickname  string                     `json:"actionUserNickname,omitempty"`
	ActionUserAvatarURL string                     `json:"actionUserAvatar,omitempty"`
	ArticleID           uint                       `json:"articleID,omitempty"`
	ArticleTitle        string                     `json:"articleTitle,omitempty"`
	CommentID           uint                       `json:"commentID,omitempty"`
	CommentContent      string                     `json:"commentContent,omitempty"`
	LinkTitle           string                     `json:"linkTitle,omitempty"`
	LinkHref            string                     `json:"linkHref,omitempty"`
	IsRead              bool                       `json:"isRead"`
	Relation            relationship_enum.Relation `json:"relation"`
}

func (NotifyApi) NotifyListView(c *gin.Context) {
	req := c.MustGet("bindReq").(NotifyListReq)
	claims := jwts.MustGetClaimsFromRequest(c)

	query := global.DB.Where("")

	// 判定返回种类
	switch req.NotifyType {
	case 1: // 评论与回复
		query = query.Where("type = ? OR type = ?", notify_enum.ArticleCommentType, notify_enum.CommentReplyType)
	case 2: // 赞和收藏
		query = query.Where("type = ? OR type = ? OR type = ?", notify_enum.ArticleLikeType, notify_enum.ArticleCollectType, notify_enum.CommentLikeType)
	case 3: // 系统通知
		query = query.Where("type = ?", notify_enum.SystemType)
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

	_list, count, err := common.ListQuery(models.NotifyModel{ReceiveUserID: claims.UserID},
		common.Options{
			PageInfo: req.PageInfo,
			Likes:    []string{"title", "content"},
			Where:    query,
			Debug:    false,
		})
	if err != nil {
		res.Fail(err, "查询数据库失败", c)
		return
	}

	var actionUserIDList []uint
	for _, model := range _list {
		if model.ActionUserID != 0 {
			actionUserIDList = append(actionUserIDList, model.ActionUserID)
		}
	}
	var m = map[uint]relationship_enum.Relation{}
	if len(actionUserIDList) > 0 {
		m = focus_service.CalcUserPatchRelationship(claims.UserID, actionUserIDList)
	}

	var list []NotifyListResp
	for _, item := range _list {
		list = append(list, NotifyListResp{
			ID:                  item.ID,
			NotifyType:          item.Type.String(),
			CreatedAt:           item.CreatedAt,
			RevUserID:           item.ReceiveUserID,
			Title:               item.Title,
			Content:             item.Content,
			ActionUserID:        item.ActionUserID,
			ActionUserNickname:  item.ActionUserNickname,
			ActionUserAvatarURL: item.ActionUserAvatarURL,
			ArticleID:           item.ArticleID,
			ArticleTitle:        item.ArticleTitle,
			CommentID:           item.CommentID,
			CommentContent:      item.CommentContent,
			LinkTitle:           item.LinkLabel,
			LinkHref:            item.LinkHref,
			IsRead:              item.IsRead,
			Relation:            m[item.ActionUserID],
		})
	}
	res.SuccessWithList(list, count, c)
}
