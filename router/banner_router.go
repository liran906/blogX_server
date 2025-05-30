// Path: ./blogX_server/router/banner_router.go

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

	r.POST("banners", mdw.BindJsonMiddleware[banner_api.BannerCreateReq], app.BannerCreateView)
	r.GET("banners", mdw.BindQueryMiddleware[banner_api.BannerListReq], mdw.AdminMiddleware, app.BannerListView)
	r.DELETE("banners", mdw.BindJsonMiddleware[models.RemoveRequest], mdw.AdminMiddleware, app.BannerRemoveView)
	r.PUT("banners/:id", mdw.AdminMiddleware, app.BannerUpdateView)
}
