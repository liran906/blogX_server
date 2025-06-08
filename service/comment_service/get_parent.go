// Path: ./service/comment_service/get_parent.go

package comment_service

import (
	"blogX_server/global"
	"blogX_server/models"
)

// GetAncestors 传入父评论 id，返回祖先评论 model 切片
func GetAncestors(parentID uint) (ancestors []models.CommentModel, err error) {
	return getAncestorsIterative(parentID)
}

// getAncestorsIterative 返回祖先 model 切片（包括入参自己）
func getAncestorsIterative(parentID uint) (ancestors []models.CommentModel, err error) {
	var cmt models.CommentModel
	if err = global.DB.Take(&cmt, parentID).Error; err != nil {
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

// getAncestorsRecursive 返回祖先 model 切片（包括入参自己, 如果要不包括, 传入 parentID 即可）
func getAncestorsRecursive(parentID uint) (ancestors []models.CommentModel, err error) {
	var cmt models.CommentModel
	if err = global.DB.Take(&cmt, parentID).Error; err != nil {
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
