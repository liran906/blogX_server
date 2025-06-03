package models

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"os"
)

type ImageModel struct {
	Model
	Filename string `gorm:"size:64; not null" json:"filename"`
	Path     string `gorm:"size:256" json:"path"`
	Url      string `gorm:"size:256" json:"url"`
	Size     int64  `gorm:"not null" json:"size"`
	Hash     string `gorm:"size:64; not null; unique" json:"hash"`
	Source   string `gorm:"size:256" json:"source"`

	// M2M
	Users []UserModel `gorm:"many2many:user_upload_images;joinForeignKey:ImageID;JoinReferences:UserID" json:"users"`
}

func (i *ImageModel) WebPath() string {
	return fmt.Sprintf("/" + i.Path) // tbd
}

// BeforeDelete 是 GORM 的钩子（Hook）方法，在记录被删除之前会自动调用
func (i *ImageModel) BeforeDelete(tx *gorm.DB) error {
	// 如果错误不是"文件不存在"，才返回错误，
	// 这样即使文件不存在，数据库记录也能正常删除。
	// 而如果是其他错误（比如权限问题），则会阻止删除操作。
	if err := os.Remove(i.Path); err != nil {
		if os.IsNotExist(err) {
			logrus.Warnf("file not exist: %s\n", err)
		} else {
			logrus.Errorf("failed to remove file: %s\n", err)
			return err
		}
	}
	return nil
}
