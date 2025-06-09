// Path: ./models/message_model.go

package models

import "blogX_server/models/enum/message_enum"

type MessageModel struct {
	Model
	Type                message_enum.Type `gorm:"not null" json:"type"`
	Title               string            `gorm:"size:128; not null" json:"title"`
	Content             string            `json:"content"`
	ReceiveUserID       uint              `gorm:"not null" json:"receiveUserID"`
	ActionUserID        uint              `json:"actionUserID"`
	ActionUserNickname  string            `gorm:"size:32" json:"actionUserNickname"`
	ActionUserAvatarURL string            `gorm:"size:256" json:"actionUserAvatarURL"`
	ArticleID           uint              `json:"articleID"`
	ArticleTitle        string            `gorm:"size:128" json:"articleTitle"`
	CommentID           uint              `json:"commentID"`
	CommentContent      string            `json:"commentContent"`
	LinkLabel           string            `gorm:"size:32" json:"linkLabel"`
	LinkHref            string            `gorm:"size:256" json:"linkHref"`
	IsRead              bool              `gorm:"not null; default:false"json:"isRead"`

	// FK
	ReceiveUserModel UserModel    `gorm:"foreignKey:ReceiveUserID; references:ID" json:"-"`
	ActionUserModel  UserModel    `gorm:"foreignKey:ActionUserID; references:ID" json:"-"`
	ArticleModel     ArticleModel `gorm:"foreignKey:ArticleID; references:ID" json:"-"`
	CommentModel     CommentModel `gorm:"foreignKey:CommentID; references:ID" json:"-"`
}
