package models

import "fmt"

type ImageModel struct {
	Model
	Filename string `gorm:"size:64; not null" json:"filename"`
	Path     string `gorm:"size:256; not null" json:"path"`
	Size     int64  `gorm:"not null" json:"size"`
	Hash     string `gorm:"size:64; not null; unique" json:"hash"`

	// M2M
	Users []UserModel `gorm:"many2many:user_upload_images;joinForeignKey:ImageID;JoinReferences:UserID" json:"users"`
}

func (i *ImageModel) WebPath() string {
	return fmt.Sprintf("/" + i.Path) // tbd
}
