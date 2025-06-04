// Path: ./common/db_transaction.go

package common

import (
	"blogX_server/global"
	"blogX_server/models"
	"gorm.io/gorm"
	"time"
)

// CreateUserAndUserConfig 如果直接用 global.DB 会导致主从数据库的 bug，主数据库没有写入，而从数据库写入了
// 所以这里用 global.DBMaster (主库)避免问题
func CreateUserAndUserConfig(u models.UserModel, uc models.UserConfigModel) (err error) {
	// 注意这里是 DBMaster
	err = global.DBMaster.Transaction(func(tx *gorm.DB) (err error) {
		// 创建 User
		err = tx.Create(&u).Error
		if err != nil {
			return err
		}

		// 设置 UserConfig 的 UserID
		uc.UserID = u.ID

		// 创建 UserConfig
		err = tx.Create(&uc).Error
		if err != nil {
			return err
		}

		// 更新 User 的 UserConfigID
		err = tx.Model(&u).Update("user_config_id", uc.UserID).Error
		if err != nil {
			return err
		}

		// 成功创建
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

// CreateUserAndUserConfig2 不走事务的存储，其实现在有专门的主库变量，现在也用不上了
func CreateUserAndUserConfig2(u models.UserModel, uc models.UserConfigModel) (err error) {
	err = global.DB.Create(&u).Error
	if err != nil {
		return err
	}

	uc.UserID = u.ID

	err = global.DB.Create(&uc).Error
	if err != nil {
		return err
	}

	err = global.DB.Model(&u).Update("user_config_id", uc.UserID).Error
	if err != nil {
		return err
	}
	return nil
}

func UpdateUserPinnedArticles(uid uint, newList []uint) error {
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
