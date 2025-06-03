// Path: ./router/banner_router.go

package router

import (
	"blogX_server/api"
	"blogX_server/api/banner_api"
	"blogX_server/middleware"
	"blogX_server/models"
	"github.com/gin-gonic/gin"
)

func BannerRouter(r *gin.RouterGroup) {
	app := api.App.BannerApi

	r.GET("banner", mdw.BindQueryMiddleware[banner_api.BannerListReq], app.BannerListView)
	r.POST("banner", mdw.BindJsonMiddleware[banner_api.BannerCreateReq], mdw.AdminMiddleware, app.BannerCreateView)
	r.DELETE("banner", mdw.BindJsonMiddleware[models.RemoveRequest], mdw.AdminMiddleware, app.BannerRemoveView)
	r.PUT("banner/:id", mdw.AdminMiddleware, app.BannerUpdateView)
}
