// Path: ./common/transaction/transaction_new_pin_aritcle.go

package transaction

import (
	"blogX_server/global"
	"blogX_server/models"
	"gorm.io/gorm"
	"strings"
)

func NewPinArticleTx(uid uint, article *models.ArticleModel) error {
	return global.DBMaster.Transaction(func(tx *gorm.DB) error {
		// 创建置顶关系
		// uid == 0 就是 admin 置顶，否则就是用户置顶
		if err := tx.Create(&models.UserPinnedArticleModel{
			UserID:    uid,
			ArticleID: article.ID,
			Rank:      1,
		}).Error; err != nil {
			if strings.Contains(err.Error(), "Duplicate entry") {
				// 已经有了，则要取消置顶
				// 修改文章 pinned 字段
				if err := tx.Where("user_id = ? AND article_id = ?", uid, article.ID).
					Delete(&models.UserPinnedArticleModel{}).Error; err != nil {
					return err
				}

				// 删除置顶关系
				if uid == 0 {
					if err := tx.Model(&models.ArticleModel{}).
						Where("id = ?", article.ID).
						Update("pinned_by_admin", false).Error; err != nil {
						return err
					}
				} else {
					if err := tx.Model(&models.ArticleModel{}).
						Where("id = ?", article.ID).
						Update("pinned_by_user", false).Error; err != nil {
						return err
					}
				}

				return nil
			}
			return err
		}

		// 修改文章 pinned 字段
		if uid == 0 {
			if err := tx.Model(&models.ArticleModel{}).
				Where("id = ?", article.ID).
				Update("pinned_by_admin", true).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Model(&models.ArticleModel{}).
				Where("id = ?", article.ID).
				Update("pinned_by_user", true).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
