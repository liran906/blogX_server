// Path: ./service/cron_service/sync_user.go

package cron_service

import (
	"blogX_server/common/transaction"
	"blogX_server/global"
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"blogX_server/service/redis_service/redis_user"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
)

const key = "homepage_visit_count"

func SyncUser() {
	// 记录时间
	now := time.Now()

	log := log_service.NewRuntimeLog("同步用户数据", log_service.RuntimeDeltaDay)
	log.SetItem("开始时间", now.Format("2006-01-02 15:04:05"))
	log.SetTitle("同步失败")

	// 备份之前的数据（如有）
	mps, err := global.Redis.HGetAll(key).Result()
	if err != nil {
		logrus.Errorf("get redis user data error: %v", err)
		log.SetItemError("查询失败", fmt.Sprintf("get redis flow data error: %v", err))
		log.SetLevel(enum.LogErrorLevel)
		log.Save()
		return
	}
	if len(mps) == 0 {
		logrus.Info("no redis user data to sync")
		log.SetTitle("无新数据")
		log.Save()
		return
	}
	err = transaction.SyncUserTx(mps)
	if err != nil {
		logrus.Errorf("sync user homepage visit counts error: %v", err)
		log.SetItemError("事务失败", fmt.Sprintf("sync user homepage visit counts error: %v", err))
		log.SetLevel(enum.LogErrorLevel)
		return
	}
	redis_user.Clear()

	logrus.Infof("update user homepage visit counts complete, %s time elapsed", time.Since(now))
	log.SetItem("完成", fmt.Sprintf("update user homepage visit counts complete, %s time elapsed", time.Since(now)))
	log.SetTitle("同步成功")
	log.Save()
}
