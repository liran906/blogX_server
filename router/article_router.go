// Path: ./router/article_router.go

package router

import (
	"blogX_server/api"
	"blogX_server/api/article_api"
	mdw "blogX_server/middleware"
	"github.com/gin-gonic/gin"
)

func ArticleRouter(rg *gin.RouterGroup) {
	app := api.App.ArticleModel

	rg.POST("article", mdw.BindJsonMiddleware[article_api.ArticleCreateReq], mdw.CaptchaMiddleware, mdw.AuthMiddleware, app.ArticleCreateView)
}
