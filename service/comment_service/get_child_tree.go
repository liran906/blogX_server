// Path: ./service/comment_service/get_child_tree.go

package comment_service

import (
	"blogX_server/global"
	"blogX_server/models"
	"time"
)

type CommentResponse struct {
	ID            uint               `json:"id"`
	CreatedAt     time.Time          `json:"createdAt"`
	Content       string             `json:"content"`
	UserID        uint               `json:"userID"`
	UserNickname  string             `json:"userNickname"`
	UserAvatarURL string             `json:"userAvatarURL"`
	ArticleID     uint               `json:"articleID"`
	ParentID      *uint              `json:"parentID"`
	LikeCount     int                `json:"likeCount"`
	ReplyCount    int                `json:"replyCount"`
	ChildComments []*CommentResponse `json:"childComments"`
}

// PreloadAllChildren 在 comment 对象的 ChildListModel 中，逐级嵌入所有 CommentModel
func PreloadAllChildren(comment *models.CommentModel) {
	global.DB.Preload("ChildListModel").Take(&comment)
	for _, child := range comment.ChildListModel {
		PreloadAllChildren(child)
	}
}

// PreloadAllChildrenResponse 返回一个 CommentResponse，其中的 ChildComments 逐级嵌入所有的 childCommentResponse
func PreloadAllChildrenResponse(cid uint) (resp *CommentResponse) {
	var comment models.CommentModel
	global.DB.Preload("UserModel").Preload("ChildListModel").Take(&comment, cid)
	
	resp = &CommentResponse{
		ID:            comment.ID,
		CreatedAt:     comment.CreatedAt,
		Content:       comment.Content,
		UserID:        comment.UserID,
		UserNickname:  comment.UserModel.Nickname,
		UserAvatarURL: comment.UserModel.AvatarURL,
		ArticleID:     comment.ArticleID,
		ParentID:      comment.ParentID,
		LikeCount:     comment.LikeCount,
		ReplyCount:    0,
		ChildComments: []*CommentResponse{},
	}
	for _, child := range comment.ChildListModel {
		resp.ChildComments = append(resp.ChildComments, PreloadAllChildrenResponse(child.ID))
	}
	return
}
