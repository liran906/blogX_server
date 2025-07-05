// Path: ./common/transaction/transaction_remove_collection.go

package transaction

import (
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/service/redis_service/redis_article"
	"blogX_server/service/redis_service/redis_cache"
	"fmt"
	"gorm.io/gorm"
)

func RemoveCollectionFolderTx(c *models.CollectionFolderModel) error {
	var relations []models.ArticleCollectionModel

	err := global.DBMaster.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("collection_folder_id = ?", c.ID).Find(&relations).Error; err != nil {
			return err
		}

		if len(relations) != 0 {
			// 收藏关系表记录删除
			if err := tx.Delete(&relations).Error; err != nil {
				return err
			}
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
	// 被收藏文章的收藏数-1（redis）在事务结束后进行
	for _, relation := range relations {
		redis_article.SubArticleCollect(relation.ArticleID)
	}
	return nil
}

func RemoveCollectionsTx(collections []models.ArticleCollectionModel) error {
	err := global.DBMaster.Transaction(func(tx *gorm.DB) error {
		// 收藏关系表记录删除
		if err := tx.Delete(&collections).Error; err != nil {
			return err
		}
		// 更新收藏夹的收藏数量
		err := tx.Model(&models.CollectionFolderModel{}).Where("id = ?", collections[0].CollectionFolderID).
			Update("article_count", gorm.Expr(fmt.Sprintf("article_count - %d", len(collections)))).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	// 被收藏文章的收藏数-1（redis）在事务结束后进行
	for _, collection := range collections {
		redis_article.SubArticleCollect(collection.ArticleID)
		redis_cache.CacheCloseCertain(fmt.Sprintf("%s%d", redis_cache.CacheArticleDetailPrefix, collection.ArticleID))
	}
	return nil
}
