// Path: ./common/transaction/transaction_remove_collection.go

package transaction

import (
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/service/redis_service/redis_article"
	"fmt"
	"gorm.io/gorm"
)

func RemoveCollection(c *models.CollectionFolderModel) error {
	var relations []models.ArticleCollectionModel
	err := global.DBMaster.Transaction(func(tx *gorm.DB) error {

		if err := tx.Where("collection_folder_id = ?", c.ID).Find(&relations).Error; err != nil {
			return err
		}

		if len(relations) == 0 {
			return fmt.Errorf("collection folder %d has no relations", c.ID)
		}

		// 收藏关系表记录删除
		if err := tx.Delete(&relations).Error; err != nil {
			return err
		}

		// 收藏夹表相关记录删除（本体）
		if err := tx.Delete(c).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	// 被收藏文章的收藏数-1（redis）
	for _, relation := range relations {
		redis_article.SubArticleCollect(relation.ArticleID)
	}
	return nil
}
