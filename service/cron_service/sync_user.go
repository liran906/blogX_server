// Path: ./service/cron_service/sync_user.go

package cron_service

import (
	"blogX_server/common/transaction"
	"blogX_server/global"
	"blogX_server/service/redis_service/redis_user"
	"github.com/sirupsen/logrus"
	"time"
)

const key = "homepage_visit_count"

func SyncUser() {
	// 记录时间
	now := time.Now()

	// 备份之前的数据（如有）
	mps, err := global.Redis.HGetAll(key).Result()
	if err != nil {
		logrus.Errorf("get redis user data error: %v", err)
		return
	}
	if len(mps) == 0 {
		logrus.Info("no redis user data to sync")
		return
	}
	err = transaction.SyncUserTx(mps)
	if err != nil {
		logrus.Errorf("sync user homepage visit counts error: %v", err)
		return
	}
	redis_user.Clear()

	logrus.Infof("update user homepage visit counts complete, %s time elapsed", time.Since(now))
}
