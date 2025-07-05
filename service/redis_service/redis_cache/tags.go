// Path: ./service/redis_service/redis_cache/tags.go

package redis_cache

import (
	"github.com/gin-gonic/gin"
	"time"
)

func NewTagsCacheOption() CacheOption {
	return CacheOption{
		Prefix: CacheTagsPrefix,
		Expiry: time.Hour,
		Params: []string{"limit", "page"},
		NoCache: func(c *gin.Context) bool {
			return false
		},
	}
}
