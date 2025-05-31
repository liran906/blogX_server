// Path: ./models/user_upload_image.go

package models

import "time"

type UserUploadImage struct {
	UserID    uint      `gorm:"primaryKey;not null;constraint:OnDelete:CASCADE" json:"userID"`  // 设置 on delete cascade
	ImageID   uint      `gorm:"primaryKey;not null;constraint:OnDelete:CASCADE" json:"imageID"` // 设置 on delete cascade
	CreatedAt time.Time `json:"createdAt"`

	// FK
	UserModel  UserModel  `gorm:"foreignKey:UserID;references:ID" json:"-"`
	ImageModel ImageModel `gorm:"foreignKey:ImageID;references:ID" json:"-"`
}
