// Path: ./blogX_server/router/mytest_router.go

package router

import (
	"blogX_server/api"
	"github.com/gin-gonic/gin"
)

func MytestRouter(rg *gin.RouterGroup) {
	app := api.App.MyTestApi

	rg.GET("test", app.MyTestView)
}
