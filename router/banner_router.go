// Path: ./blogX_server/router/banner_router.go

package router

import (
	"blogX_server/api"
	"blogX_server/middleware"
	"github.com/gin-gonic/gin"
)

func BannerRouter(r *gin.RouterGroup) {
	app := api.App.BannerApi

	r.POST("banners", app.BannerCreateView)
	r.GET("banners", middleware.AdminMiddleware, app.BannerListView)
	r.DELETE("banners", middleware.AdminMiddleware, app.BannerRemoveView)
	r.PUT("banners/:id", middleware.AdminMiddleware, app.BannerUpdateView)
}
