// Path: ./models/comment_likes_model.go

package models

import "time"

// CommentLikesModel 评论点赞表
type CommentLikesModel struct {
	UserID    uint      `gorm:"primaryKey" json:"userID"`
	CommentID uint      `gorm:"primaryKey" json:"commentID"`
	CreatedAt time.Time `json:"createdAt"`

	// FK
	UserModel    UserModel    `gorm:"foreignKey:UserID;references:ID" json:"-"`
	CommentModel CommentModel `gorm:"foreignKey:CommentID;references:ID" json:"-"`
}
