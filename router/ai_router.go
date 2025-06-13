// Path: ./router/ai_router.go

package router

import (
	"blogX_server/api"
	"blogX_server/api/ai_api"
	mdw "blogX_server/middleware"
	"github.com/gin-gonic/gin"
)

func AIRouter(rg *gin.RouterGroup) {
	app := api.App.AiApi

	rg.POST("ai", mdw.BindJsonMiddleware[ai_api.ArticleAnalysisReq], mdw.AuthMiddleware, app.ArticleAnalysisView)
	rg.GET("ai_search", mdw.BindQueryMiddleware[ai_api.ArticleAiReq], mdw.AuthMiddleware, app.ArticleAiView)
}
