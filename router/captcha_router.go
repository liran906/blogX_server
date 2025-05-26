// Path: ./blogX_server/router/captcha_router.go

package router

import (
	"blogX_server/api"
	"github.com/gin-gonic/gin"
)

func CaptchaRouter(rg *gin.RouterGroup) {
	app := api.App.CaptchaApi

	rg.GET("captcha", app.CaptchaView)
}
