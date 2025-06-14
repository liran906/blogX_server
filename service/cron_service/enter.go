// Path: ./service/cron_service/enter.go

package cron_service

import (
	"blogX_server/global"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

type CronService struct{}

func Cron() {
	timezone, _ := time.LoadLocation("Asia/Shanghai")
	crontab := cron.New(cron.WithSeconds(), cron.WithLocation(timezone))

	// 每天固定时间 去同步文章数据
	_, err1 := crontab.AddFunc(global.Config.Redis.ArticleSyncTime, SyncArticle)
	_, err2 := crontab.AddFunc(global.Config.Redis.CommentSyncTime, SyncComment)
	_, err3 := crontab.AddFunc(global.Config.Redis.SiteDataSyncTime, SyncData)
	_, err4 := crontab.AddFunc(global.Config.Redis.UserDataSyncTime, SyncUser)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		logrus.Panicln("crontab.AddFunc err:", err1)
		logrus.Panicln("crontab.AddFunc err:", err2)
		logrus.Panicln("crontab.AddFunc err:", err3)
		logrus.Panicln("crontab.AddFunc err:", err4)
		return
	}
	crontab.Start()
}

func mapKeys(m map[uint]struct{}) []uint {
	keys := make([]uint, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func mapMergeAndConvert(base, delta map[uint]int) map[string]interface{} {
	merged := make(map[uint]int, len(base)+len(delta))

	// 始终复制 base
	for k, v := range base {
		merged[k] = v
	}

	// 合并 delta（如果存在）
	for k, v := range delta {
		merged[k] += v
	}

	// 构建 Redis-friendly map
	result := make(map[string]interface{}, len(merged))
	for k, v := range merged {
		result[strconv.Itoa(int(k))] = v
	}
	return result
}
