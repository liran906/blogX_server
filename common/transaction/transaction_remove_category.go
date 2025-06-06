// Path: ./common/transaction/transaction_remove_category.go

package transaction

import (
	"blogX_server/global"
	"blogX_server/models"
	"gorm.io/gorm"
)

func RemoveCategory(c *models.CategoryModel) error {
	return global.DBMaster.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(c).Association("ArticleList").Find(&c.ArticleList); err != nil {
			return err
		}
		for _, a := range c.ArticleList {
			if err := tx.Model(&a).Update("CategoryID", nil).Error; err != nil {
				return err
			}
		}
		err := tx.Delete(c).Error
		if err != nil {
			return err
		}
		return nil
	})
}
