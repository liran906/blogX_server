// Path: ./service/redis_service/redis_cache/enter.go

package redis_cache

import (
	"blogX_server/global"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"time"
)

type CacheOption struct {
	Prefix  CacheMiddlewarePrefix
	Expiry  time.Duration
	Params  []string // 用于 Query 参数
	NoCache func(c *gin.Context) bool
	IsUri   bool
	UriKey  string // 用于 URI 参数，如 ":id"
}

type CacheMiddlewarePrefix string

const (
	CacheBannerPrefix        CacheMiddlewarePrefix = "cache_banner_"
	CacheTagsPrefix          CacheMiddlewarePrefix = "cache_tags_"
	CacheArticleDetailPrefix CacheMiddlewarePrefix = "cache_article_detail_"
)

func CacheOpen(key, value string, expiry time.Duration) {
	global.Redis.Set(key, value, expiry)
}

func CacheCloseAll(prefix CacheMiddlewarePrefix) {
	keys, err := global.Redis.Keys(fmt.Sprintf("%s*", prefix)).Result()
	if err != nil {
		logrus.Errorf(err.Error())
		return
	}
	if len(keys) > 0 {
		logrus.Infof("删除前缀 %s 缓存 共 %d 条", prefix, len(keys))
		global.Redis.Del(keys...)
	}
}

func CacheCloseCertain(key string) {
	global.Redis.Del(key)
}
