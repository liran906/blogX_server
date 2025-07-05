// Path: ./service/redis_service/redis_cache/article_detail.go

package redis_cache

import (
	"github.com/gin-gonic/gin"
	"time"
)

func NewArticleDetailCacheOption() CacheOption {
	return CacheOption{
		Prefix: CacheArticleDetailPrefix,
		Expiry: time.Hour,
		IsUri:  true,
		UriKey: "id",
		NoCache: func(c *gin.Context) bool {
			return false
		},
	}
}
