// Path: ./service/comment_service/get_child_tree.go

package comment_service

import (
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/service/redis_service/redis_comment"
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
	IsLiked       bool               `json:"isLiked"`
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
func PreloadAllChildrenResponseFromID(cid uint, userLikeMap map[uint]struct{}) (resp *CommentResponse) {
	var cmt models.CommentModel
	global.DB.Preload("UserModel").Preload("ChildListModel").Take(&cmt, cid)
	return PreloadAllChildrenResponseFromModel(&cmt, userLikeMap)
}

func PreloadAllChildrenResponseFromModel(cmt *models.CommentModel, userLikeMap map[uint]struct{}) (resp *CommentResponse) {
	// 两种实现，功能完全一致

	// preloadByAttr 是更简单的实现，但需要维护 depth 属性
	return preloadByAttr(cmt, userLikeMap)

	// preloadByClosure 利用闭包（depth 的递归和回溯）判断是否达到层数限制
	//return preloadByClosure(cmt)
}

// preloadByClosure 利用闭包（depth 的递归和回溯）判断是否达到层数限制
func preloadByClosure(cmt *models.CommentModel, userLikeMap map[uint]struct{}) (resp *CommentResponse) {
	var depth int
	// 也可以 helper 函数多传入一个 depth 参数，这样不需要回溯了
	var preloadHelper func(*models.CommentModel) *CommentResponse
	preloadHelper = func(cmt *models.CommentModel) (resp *CommentResponse) {
		depth++
		if depth > global.Config.Site.Article.CommentDepth {
			depth--
			return
		}
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
			LikeCount:     cmt.LikeCount + redis_comment.GetCommentLikeCount(cmt.ID),
			ReplyCount:    cmt.ReplyCount + redis_comment.GetCommentReplyCount(cmt.ID),
			ChildComments: []*CommentResponse{},
		}
		// 判断是否被本人点过赞
		if _, ok := userLikeMap[cmt.ID]; ok {
			resp.IsLiked = true
		}
		for i := range cmt.ChildListModel {
			child := cmt.ChildListModel[i]
			resp.ChildComments = append(resp.ChildComments, preloadHelper(child))
		}
		depth--
		return
	}

	return preloadHelper(cmt)
}

// preloadByAttr 是更简单的实现，但需要维护 depth 属性
func preloadByAttr(cmt *models.CommentModel, userLikeMap map[uint]struct{}) (resp *CommentResponse) {
	if cmt.Depth >= global.Config.Site.Article.CommentDepth {
		return
	}
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
		LikeCount:     cmt.LikeCount + redis_comment.GetCommentLikeCount(cmt.ID),
		ReplyCount:    cmt.ReplyCount + redis_comment.GetCommentReplyCount(cmt.ID),
		ChildComments: []*CommentResponse{},
	}
	// 判断是否被本人点过赞
	if _, exists := userLikeMap[cmt.ID]; exists {
		resp.IsLiked = true
	}
	for i := range cmt.ChildListModel {
		child := cmt.ChildListModel[i]
		resp.ChildComments = append(resp.ChildComments, preloadByAttr(child, userLikeMap))
	}
	return
}
