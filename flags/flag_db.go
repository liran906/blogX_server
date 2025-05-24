package flags

import (
	"blogX_server/global"
	"blogX_server/models"
	"github.com/sirupsen/logrus"
)

func FlagDB() {
	err := global.DB.AutoMigrate(
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
