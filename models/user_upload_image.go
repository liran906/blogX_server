// Path: ./blogX_server/models/user_upload_image.go

package models

import "time"

type UserUploadImage struct {
	UserID    uint      `gorm:"primaryKey" json:"userID"`
	ImageID   uint      `gorm:"primaryKey" json:"imageID"`
	CreatedAt time.Time `json:"createdAt"`

	// FK
	UserModel  UserModel  `gorm:"foreignKey:UserID;references:ID" json:"-"`
	ImageModel ImageModel `gorm:"foreignKey:ImageID;references:ID" json:"-"`
}
