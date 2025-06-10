// Path: ./common/transaction/transaction_remove_global_notification.go

package transaction

import (
	"blogX_server/global"
	"blogX_server/models"
	"gorm.io/gorm"
)

func RemoveGlobalNotificationTx(gn *models.GlobalNotificationModel) error {
	return global.DB.Transaction(func(tx *gorm.DB) (err error) {
		// 删除本体
		if err = tx.Delete(gn).Error; err != nil {
			return err
		}
		// 删除用户全局表中的消息
		err = tx.Where("global_notification_id = ?", gn.ID).Delete(&models.UserGlobalNotificationModel{}).Error
		if err != nil {
			return err
		}
		return nil
	})
}
