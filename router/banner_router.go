// Path: ./blogX_server/router/banner_router.go

package router

import (
	"blogX_server/api"
	"blogX_server/middleware"
	"github.com/gin-gonic/gin"
)

func BannerRouter(rg *gin.RouterGroup) {
	// 绑定中间件，注意不能直接在传入的指针上使用，否则其他视图都会被绑定
	r := rg.Group("").Use(middleware.AdminMiddleware)

	app := api.App.BannerApi

	r.POST("banners", app.BannerCreateView)
	r.GET("banners", app.BannerListView)
	r.DELETE("banners", app.BannerRemoveView)
	r.PUT("banners", app.BannerUpdateView)
}
