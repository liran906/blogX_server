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
	RootID        *uint              `json:"rootID"`
	Depth         int                `json:"depth"`
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

// PreloadAllChildrenResponseFromID 返回一个 CommentResponse，其中的 ChildComments 逐级嵌入所有的 childCommentResponse
func PreloadAllChildrenResponseFromID(cid uint) (resp *CommentResponse) {
	var cmt models.CommentModel
	global.DB.Preload("UserModel").Preload("ChildListModel").Take(&cmt, cid)
	return PreloadAllChildrenResponseFromModel(&cmt)
}

func PreloadAllChildrenResponseFromModel(cmt *models.CommentModel) (resp *CommentResponse) {
	global.DB.Preload("UserModel").Preload("ChildListModel").Take(cmt)
	resp = &CommentResponse{
		ID:            cmt.ID,
		CreatedAt:     cmt.CreatedAt,
		Content:       cmt.Content,
		UserID:        cmt.UserID,
		UserNickname:  cmt.UserModel.Nickname,
		UserAvatarURL: cmt.UserModel.AvatarURL,
		ArticleID:     cmt.ArticleID,
		ParentID:      cmt.ParentID,
		RootID:        cmt.RootID,
		Depth:         cmt.Depth,
		LikeCount:     cmt.LikeCount,
		ReplyCount:    len(cmt.ChildListModel),
		ChildComments: []*CommentResponse{},
	}
	for i := range cmt.ChildListModel {
		child := cmt.ChildListModel[i]
		resp.ChildComments = append(resp.ChildComments, PreloadAllChildrenResponseFromModel(child))
	}
	return
}
