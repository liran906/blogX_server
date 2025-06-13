// Path: ./middleware/increase_site_click_middleware.go

package mdw

import (
	"blogX_server/service/redis_service/redis_site"
	"github.com/gin-gonic/gin"
)

func IncreaseSiteClickMiddleware(c *gin.Context) {
	redis_site.IncreaseClick()
}
