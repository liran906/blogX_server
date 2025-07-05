// Path: ./service/redis_service/redis_cache/banner.go

package redis_cache

import (
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

func NewBannerCacheOption() CacheOption {
	return CacheOption{
		Prefix: CacheBannerPrefix,
		Expiry: time.Hour,
		Params: []string{"type"},
		NoCache: func(c *gin.Context) bool {
			var referer = c.GetHeader("referer")
			if strings.Contains(referer, "admin") {
				// 后台来的，不走缓存
				return true
			}
			return false
		},
	}
}
