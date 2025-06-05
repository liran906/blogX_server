// Path: ./service/redis_service/redis_article/user_article_history.go

package redis_article

import (
	"blogX_server/global"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
)

func setUserArticleHistoryCache(articleID, userId uint) {
	key := fmt.Sprintf("history_%d", userId)
	field := fmt.Sprintf("a_%d", articleID)
	err := global.Redis.HSet(key, field, "").Err()
	if err != nil {
		logrus.Error("Redis Set error: " + err.Error())
		return
	}
}

func SetUserArticleHistoryCacheToday(articleID, userId uint) {
	if HasUserArticleHistoryCacheToday(articleID, userId) {
		setUserArticleHistoryCache(articleID, userId)
	} else {
		setUserArticleHistoryCache(articleID, userId)

		tomorrow := time.Now().AddDate(0, 0, 1)                                                     // 获取今天剩余时间
		end := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, time.Local) // 注意设置时区
		key := fmt.Sprintf("history_%d", userId)                                                    // 读取 key
		global.Redis.ExpireAt(key, end)                                                             // 设置 redis 过期
	}
}

func HasUserArticleHistoryCacheToday(articleID, userId uint) (ok bool) {
	key := fmt.Sprintf("history_%d", userId)
	field := fmt.Sprintf("a_%d", articleID)

	exists, err := global.Redis.HExists(key, field).Result()
	if err != nil {
		logrus.Error("Redis Exists error: " + err.Error())
		return
	}
	return exists
}
