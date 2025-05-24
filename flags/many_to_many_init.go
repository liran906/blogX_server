// Path: ./blogX_server/flags/many_to_many_init.go

package flags

import (
	"blogX_server/global"
	"blogX_server/models"
	"github.com/sirupsen/logrus"
)

func ManyToManyInit() error {
	// user_upload_image
	err := global.DB.SetupJoinTable(&models.UserModel{}, "Images", &models.UserUploadImage{})
	if err != nil {
		logrus.Errorf("failed to setup join table (%s): %s\n", "user_upload_image_user", err)
		return err
	}
	err = global.DB.SetupJoinTable(&models.ImageModel{}, "Users", &models.UserUploadImage{})
	if err != nil {
		logrus.Errorf("failed to setup join table (%s): %s\n", "user_upload_image_image", err)
		return err
	}
	return nil
}
