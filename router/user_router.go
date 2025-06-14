// Path: ./router/user_router.go

package router

import (
	"blogX_server/api"
	"blogX_server/api/user_api"
	"blogX_server/middleware"
	"blogX_server/models"
	"github.com/gin-gonic/gin"
)

func UserRouter(rg *gin.RouterGroup) {
	app := api.App.UserApi

	rg.POST("user/send_email", mdw.BindJsonMiddleware[user_api.SendEmailReq], mdw.CaptchaMiddleware, app.SendEmailView)
	rg.POST("user/register_email", mdw.BindJsonMiddleware[user_api.RegisterEmailReq], mdw.EmailRegisterMiddleware, mdw.CaptchaMiddleware, mdw.EmailVerifyMiddleware, mdw.RegisterVerifyMiddleware, app.RegisterEmailView)
	rg.POST("user/login", mdw.BindJsonMiddleware[user_api.PwdLoginReq], mdw.UsernamePwdLoginMiddleware, mdw.CaptchaMiddleware, app.PwdLoginView)
	rg.GET("user/:id", mdw.BindUriMiddleware[models.IDRequest], mdw.AuthMiddleware, app.UserDetailView)
	rg.GET("user", mdw.BindQueryMiddleware[models.IDRequest], app.UserBriefInfoView)
	rg.GET("user/login_list", mdw.BindQueryMiddleware[user_api.UserLoginListReq], mdw.AuthMiddleware, app.UserLoginListView)
	rg.GET("user/list", mdw.BindQueryMiddleware[user_api.UserListReq], mdw.AdminMiddleware, app.UserListView)
	rg.PUT("user/pwd", mdw.BindJsonMiddleware[user_api.ChangePasswordReq], mdw.AuthMiddleware, app.ChangePasswordView)
	rg.PUT("user/pwd/reset", mdw.BindJsonMiddleware[user_api.ResetPasswordReq], mdw.CaptchaMiddleware, mdw.EmailVerifyMiddleware, app.ResetPasswordView)
	rg.PUT("user/bind_email", mdw.CaptchaMiddleware, mdw.AuthMiddleware, mdw.EmailVerifyMiddleware, app.BindEmailView)
	rg.PUT("user/update", mdw.BindJsonMiddleware[user_api.UserInfoUpdateReq], mdw.AuthMiddleware, app.UserInfoUpdateView)
	rg.PUT("user/update/admin", mdw.BindJsonMiddleware[user_api.AdminUpdateUserReq], mdw.AdminMiddleware, app.AdminUpdateUserView)
	rg.DELETE("user/logout", mdw.AuthMiddleware, app.UserLogoutView)
}
