// Path: ./models/user_config_model.go

package models

import "time"

type UserConfigModel struct {
	UserID             uint       `gorm:"primaryKey" json:"userID"`
	UpdatedAt          *time.Time `json:"updatedAt"`                                        // 上次修改时间，可能为空，所以是指针
	Tags               []string   `gorm:"type:longtext; serializer:json" json:"tags"`       // 兴趣标签
	ThemeID            uint8      `gorm:"not null; default:1" json:"themeID"`               // 主页样式 id
	DisplayCollections bool       `gorm:"not null; default:true" json:"displayCollections"` // 公开我的收藏
	DisplayFans        bool       `gorm:"not null; default:true" json:"displayFans"`        // 公开我的粉丝
	DisplayFollowing   bool       `gorm:"not null; default:true" json:"displayFollowing"`   // 公开我的关注
	HomepageVisitCount int        `gorm:"not null; default:0" json:"homepageVisitCount"`    // 主页访问量

	// FK
	UserModel UserModel `gorm:"foreignKey:UserID;references:ID" json:"userModel"` // 外键关联到 User, ref 如果不写会自动关联到 ID
}

type UserMessageConfModel struct {
	UserID                 uint `gorm:"primary_key" json:"userID"`
	ReceiveCommentNotify   bool `gorm:"not null; default:true" json:"receiveCommentNotify"`
	ReceiveLikeNotify      bool `gorm:"not null; default:true" json:"receiveLikeNotify"`
	ReceiveCollectNotify   bool `gorm:"not null; default:true" json:"receiveCollectNotify"`
	ReceivePrivateMessage  bool `gorm:"not null; default:true" json:"receivePrivateMessage"`
	ReceiveStrangerMessage bool `gorm:"not null; default:true" json:"receiveStrangerMessage"`

	// FK
	UserModel UserModel `gorm:"foreignKey:UserID; reference:ID" json:"userModel"`
}
