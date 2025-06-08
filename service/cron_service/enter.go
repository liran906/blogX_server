// Path: ./service/cron_service/enter.go

package cron_service

import (
	"blogX_server/global"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"time"
)

type CronService struct{}

func Cron() {
	timezone, _ := time.LoadLocation("Asia/Shanghai")
	crontab := cron.New(cron.WithSeconds(), cron.WithLocation(timezone))

	// 每天固定时间 去同步文章数据
	_, err1 := crontab.AddFunc(global.Config.Redis.ArticleSyncTime, SyncArticle)
	_, err2 := crontab.AddFunc(global.Config.Redis.ArticleSyncTime, SyncComment)
	if err1 != nil || err2 != nil {
		logrus.Panicln("crontab.AddFunc err:", err1)
		logrus.Panicln("crontab.AddFunc err:", err2)
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
