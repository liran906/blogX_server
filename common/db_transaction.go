// Path: ./common/db_transaction.go

package common

import (
	"blogX_server/global"
	"blogX_server/models"
	"gorm.io/gorm"
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
