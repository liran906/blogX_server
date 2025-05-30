// Path: ./blogX_server/router/image_router.go

package router

import (
	"blogX_server/api"
	"blogX_server/middleware"
	"github.com/gin-gonic/gin"
)

func ImageRouter(r *gin.RouterGroup) {
	app := api.App.ImageApi

	r.POST("images", mdw.AuthMiddleware, app.ImageUploadView)
	r.GET("images", mdw.AdminMiddleware, app.ImageListView)
	r.DELETE("images", mdw.AdminMiddleware, app.ImageRemoveView)
}
