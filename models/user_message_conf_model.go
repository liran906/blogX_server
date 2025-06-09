// Path: ./models/user_message_conf_model.go

package models

type UserMessageConfModel struct {
	UserID                 uint `gorm:"primary_key" json:"userID"`
	ReceiveCommentMessage  bool `gorm:"not null; default:true" json:"receiveCommentMessage"`
	ReceiveLikeMessage     bool `gorm:"not null; default:true" json:"receiveLikeMessage"`
	ReceiveCollectMessage  bool `gorm:"not null; default:true" json:"receiveCollectMessage"`
	ReceivePrivateMessage  bool `gorm:"not null; default:true" json:"receivePrivateMessage"`
	ReceiveStrangerMessage bool `gorm:"not null; default:true" json:"receiveStrangerMessage"`

	// FK
	UserModel UserModel `gorm:"foreignKey:UserID; reference:ID" json:"userModel"`
}
