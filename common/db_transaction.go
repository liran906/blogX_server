// Path: ./blogX_server/common/db_transaction.go

package common

import (
	"blogX_server/global"
	"blogX_server/models"
	"gorm.io/gorm"
)

func CreateUserAndUserConfig(u models.UserModel, uc models.UserConfigModel) (err error) {
	err = global.DB.Transaction(func(tx *gorm.DB) (err error) {
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
