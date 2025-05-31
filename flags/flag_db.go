// Path: ./flags/flag_db.go

package flags

import (
	"blogX_server/global"
	"blogX_server/models"
	"github.com/sirupsen/logrus"
)

func FlagDB() {
	// 映射自定义多对多关系
	err := ManyToManyInit()
	if err != nil {
		return
	}

	// 表迁移
	err = global.DB.AutoMigrate(
		&models.UserModel{},
		&models.UserConfigModel{},
		&models.ArticleModel{},
		&models.CategoryModel{},
		&models.ArticleLikesModel{},
		&models.CollectionFolderModel{},
		&models.ArticleCollectionModel{},
		&models.UserPinnedArticleModel{},
		&models.ImageModel{},
		&models.UserArticleHistoryModel{},
		&models.CommentModel{},
		&models.BannerModel{},
		&models.LogModel{},
		&models.UserLoginModel{},
		&models.GlobalNotificationModel{},
		&models.UserUploadImage{},
	)
	if err != nil {
		logrus.Errorf("failed to migrate DB: %s\n", err)
		return
	}
	logrus.Info("DB migration successful")
}

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
