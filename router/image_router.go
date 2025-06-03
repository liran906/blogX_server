// Path: ./router/image_router.go

package router

import (
	"blogX_server/api"
	"blogX_server/api/image_api"
	"blogX_server/middleware"
	"blogX_server/models"
	"github.com/gin-gonic/gin"
)

func ImageRouter(r *gin.RouterGroup) {
	app := api.App.ImageApi

	r.POST("images", mdw.AuthMiddleware, app.ImageUploadView)
	r.POST("images/cache", mdw.BindJsonMiddleware[image_api.ImageCacheReq], mdw.AuthMiddleware, app.ImageCacheView)
	r.GET("images", mdw.BindQueryMiddleware[image_api.ImageListReq], mdw.AdminMiddleware, app.ImageListView)
	r.DELETE("images", mdw.BindJsonMiddleware[models.RemoveRequest], mdw.AdminMiddleware, app.ImageRemoveView)
}
