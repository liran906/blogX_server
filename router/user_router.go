// Path: ./blogX_server/router/user_router.go

package router

import (
	"blogX_server/api"
	"blogX_server/middleware"
	"github.com/gin-gonic/gin"
)

func UserRouter(rg *gin.RouterGroup) {
	app := api.App.UserApi

	rg.POST("user/send_email", middleware.CaptchaMiddleware, app.SendEmailView)
}
