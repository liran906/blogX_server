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

	// 每天2点去同步文章数据
	_, err := crontab.AddFunc(global.Config.Redis.SyncTime, SyncArticle)
	if err != nil {
		logrus.Panicln("crontab.AddFunc err:", err)
		return
	}

	crontab.Start()
}
