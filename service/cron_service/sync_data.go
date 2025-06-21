// Path: ./service/cron_service/sync_data.go

package cron_service

import (
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"blogX_server/service/redis_service/redis_site"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
)

const flowKey = "blogx_site_flow"
const clickKey = "blogx_site_click"

func SyncData() {
	// 记录时间
	now := time.Now()

	log := log_service.NewRuntimeLog("同步站点数据", log_service.RuntimeDeltaDay)
	log.SetItem("开始时间", now.Format("2006-01-02 15:04:05"))
	log.SetTitle("同步失败")

	// 备份之前的数据（如有）
	mps, err := global.Redis.HGetAll(flowKey).Result()
	if err != nil {
		logrus.Errorf("get redis flow data error: %v", err)
		log.SetItemError("查询失败", fmt.Sprintf("get redis flow data error: %v", err))
		log.SetLevel(enum.LogErrorLevel)
		log.Save()
		return
	}
	for k := range mps {
		if k == now.Format("2006-01-02") {
			continue // 遇到今天不同步（数据还不完整）
		}
		err = sync(k)
		if err == nil {
			global.Redis.HDel(flowKey, k)
			global.Redis.HDel(clickKey, k)
		}
	}
	logrus.Infof("update site statistics complete, %s time elapsed", time.Since(now))
	log.SetItem("完成", fmt.Sprintf("update site statistics complete, %s time elapsed", time.Since(now)))
	log.SetTitle("同步成功")
	log.Save()
}

func sync(field string) error {
	// 从 redis 中读取数据
	flow := redis_site.GetFlow(field)
	click := redis_site.GetClick(field)
	date, _ := time.ParseInLocation("2006-01-02", field, time.Local)
	date = date.Add(12 * time.Hour) // 这里加 12 小时，方便对比

	err := global.DB.Create(&models.DataModel{
		Date:       date,
		FlowCount:  flow,
		ClickCount: click,
	}).Error
	if err != nil {
		return err
	}

	redis_site.ClearAllByDate(field)
	return nil
}
