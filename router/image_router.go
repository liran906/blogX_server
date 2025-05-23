// Path: ./blogX_server/router/image_router.go

package router

import (
	"blogX_server/api"
	"blogX_server/middleware"
	"github.com/gin-gonic/gin"
)

func ImageRouter(r *gin.RouterGroup) {
	app := api.App.ImageAip

	r.POST("images", middleware.AuthMiddleware, app.ImageUploadView)
}
