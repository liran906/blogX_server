// Path: ./service/comment_service/get_root.go

package comment_service

import (
	"blogX_server/global"
	"blogX_server/models"
)

func GetRootCommentID(parentID *uint) (rootID *uint, err error) {
	if parentID == nil {
		return
	}
	var parent models.CommentModel
	err = global.DB.Take(&parent, parentID).Error
	if err != nil {
		return
	}
	if parent.RootID == nil {
		return &parent.ID, nil
	}
	return parent.RootID, nil
}
