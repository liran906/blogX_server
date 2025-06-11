// Path: ./router/search_router.go

package router

import (
	"blogX_server/api"
	"blogX_server/api/search_api"
	mdw "blogX_server/middleware"
	"github.com/gin-gonic/gin"
)

func SearchRouter(rg *gin.RouterGroup) {
	app := api.App.SearchApi

	rg.GET("search/article", mdw.BindQueryMiddleware[search_api.ArticleSearchReq], app.ArticleSearchView)
}
