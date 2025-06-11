// Path: ./common/transaction/transcation_pin_article.go

package transaction

import (
	"blogX_server/global"
	"blogX_server/models"
	"gorm.io/gorm"
	"time"
)

func UpdateUserPinnedArticlesTx(uid uint, newList []uint) error {
	return global.DBMaster.Transaction(func(tx *gorm.DB) error {
		var oldPinned []models.UserPinnedArticleModel
		if err := tx.Where("user_id = ?", uid).Find(&oldPinned).Error; err != nil {
			return err
		}

		if len(oldPinned) > 0 {
			// 收集旧的文章 ID
			var oldList []uint
			for _, item := range oldPinned {
				oldList = append(oldList, item.ArticleID)
			}

			// 新旧相等
			if equalUintSets(oldList, newList) {
				return nil
			}

			// 取消旧置顶标志
			if err := tx.Model(&models.ArticleModel{}).
				Where("id IN ?", oldList).
				Update("pinned_by_user", false).Error; err != nil {
				return err
			}

			// 删除旧记录
			if err := tx.Where("user_id = ?", uid).
				Delete(&models.UserPinnedArticleModel{}).Error; err != nil {
				return err
			}
		}

		if len(newList) > 0 {
			// 插入新置顶记录（推荐批量）
			var records []models.UserPinnedArticleModel
			now := time.Now()
			for i, aid := range newList {
				records = append(records, models.UserPinnedArticleModel{
					UserID:    uid,
					ArticleID: aid,
					Rank:      i + 1,
					CreatedAt: now,
				})
			}

			if err := tx.Create(&records).Error; err != nil {
				return err
			}

			// 更新新置顶标志
			if err := tx.Model(&models.ArticleModel{}).
				Where("id IN ?", newList).
				Update("pinned_by_user", true).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// UpdateSitePinnedArticlesTx 整体更新站点置顶文章
// 这里还是用的 UserPinnedArticleModel 这个表来记录
// 但是 uid 设为 0
func UpdateSitePinnedArticlesTx(newList []uint) error {
	return global.DBMaster.Transaction(func(tx *gorm.DB) error {
		var oldPinned []models.UserPinnedArticleModel
		if err := tx.Where("user_id = ?", 0).Find(&oldPinned).Error; err != nil {
			return err
		}

		if len(oldPinned) > 0 {
			// 收集旧的文章 ID
			var oldList []uint
			for _, item := range oldPinned {
				oldList = append(oldList, item.ArticleID)
			}

			// 新旧相等
			if equalUintSets(oldList, newList) {
				return nil
			}

			// 取消旧置顶标志
			if err := tx.Model(&models.ArticleModel{}).
				Where("id IN ?", oldList).
				Update("pinned_by_admin", false).Error; err != nil {
				return err
			}

			// 删除旧记录
			if err := tx.Where("user_id = ?", 0).
				Delete(&models.UserPinnedArticleModel{}).Error; err != nil {
				return err
			}
		}

		if len(newList) > 0 {
			// 插入新置顶记录（推荐批量）
			var records []models.UserPinnedArticleModel
			now := time.Now()
			for i, aid := range newList {
				records = append(records, models.UserPinnedArticleModel{
					UserID:    0,
					ArticleID: aid,
					Rank:      i + 1,
					CreatedAt: now,
				})
			}

			if err := tx.Create(&records).Error; err != nil {
				return err
			}

			// 更新新置顶标志
			if err := tx.Model(&models.ArticleModel{}).
				Where("id IN ?", newList).
				Update("pinned_by_admin", true).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// equalUintSets 判断是否一样（无序）
func equalUintSets(a, b []uint) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[uint]int)
	for _, v := range a {
		m[v]++
	}
	for _, v := range b {
		if m[v] == 0 {
			return false
		}
		m[v]--
	}
	return true
}
