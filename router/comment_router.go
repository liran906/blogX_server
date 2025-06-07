// Path: ./router/comment_router.go

package router

import (
	"blogX_server/api"
	"blogX_server/api/comment_api"
	mdw "blogX_server/middleware"
	"blogX_server/models"
	"github.com/gin-gonic/gin"
)

func CommentRouter(rg *gin.RouterGroup) {
	app := api.App.CommentApi

	rg.POST("comment", mdw.BindJsonMiddleware[comment_api.CommentCreateReq], mdw.AuthMiddleware, app.CommentCreateView)
	rg.GET("article/:id/comment", mdw.BindUriMiddleware[models.IDRequest], app.CommentTreeView)
}
