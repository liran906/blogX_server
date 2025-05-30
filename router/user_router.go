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
	rg.POST("user/register_email", middleware.EmailRegisterMiddleware, middleware.CaptchaMiddleware, middleware.EmailVerifyMiddleware, middleware.RegisterVerifyMiddleware, app.RegisterEmailView)
	rg.POST("user/login", middleware.UsernamePwdLoginMiddleware, middleware.CaptchaMiddleware, app.PwdLoginView)
	rg.GET("user/:id", middleware.AuthMiddleware, app.UserDetailView)
	rg.GET("user", app.UserBriefInfoView)
	rg.GET("user/list", middleware.AuthMiddleware, app.UserLoginListView)
	rg.PUT("user/pwd", middleware.AuthMiddleware, app.ChangePasswordView)
	rg.PUT("user/pwd/reset", middleware.CaptchaMiddleware, middleware.EmailVerifyMiddleware, app.ResetPasswordView)
	rg.PUT("user/bind_email", middleware.CaptchaMiddleware, middleware.AuthMiddleware, middleware.EmailVerifyMiddleware, app.BindEmailView)
	rg.PUT("user/update", middleware.AuthMiddleware, app.UserInfoUpdateView)
}
