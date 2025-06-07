// Path: ./service/comment_service/get_parent.go

package comment_service

import (
	"blogX_server/global"
	"blogX_server/models"
)

func GetAncestors(commentID uint) (ancestors []models.CommentModel, err error) {
	return getAncestorsIterative(commentID)
}

func getAncestorsIterative(commentID uint) (ancestors []models.CommentModel, err error) {
	var cmt models.CommentModel
	if err = global.DB.Take(&cmt, commentID).Error; err != nil {
		return
	}
	for cmt.ParentID != nil {
		ancestors = append(ancestors, cmt)
		var parent models.CommentModel
		if err = global.DB.Take(&parent, cmt.ParentID).Error; err != nil {
			return
		}
		cmt = parent
	}
	ancestors = append(ancestors, cmt)
	return
}

func getAncestorsRecursive(commentID uint) (ancestors []models.CommentModel, err error) {
	var cmt models.CommentModel
	if err = global.DB.Take(&cmt, commentID).Error; err != nil {
		return
	}
	if cmt.ParentID == nil {
		ancestors = append(ancestors, cmt)
		return
	}
	ancestors = append(ancestors, cmt)
	ans, err := getAncestorsRecursive(*cmt.ParentID)
	if err != nil {
		return
	}
	ancestors = append(ancestors, ans...)
	return
}
