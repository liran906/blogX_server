// Path: ./router/search_router.go

package router

import (
	"blogX_server/api"
	"blogX_server/api/search_api"
	"blogX_server/common"
	mdw "blogX_server/middleware"
	"blogX_server/service/redis_service/redis_cache"
	"github.com/gin-gonic/gin"
)

func SearchRouter(rg *gin.RouterGroup) {
	app := api.App.SearchApi

	rg.GET("search/article", mdw.BindQueryMiddleware[search_api.ArticleSearchReq], app.ArticleSearchView)
	rg.GET("search/text", mdw.BindQueryMiddleware[search_api.TextSearchReq], app.TextSearchView)
	rg.GET("search/tags", mdw.BindQueryMiddleware[common.PageInfo], mdw.CacheMiddleware(redis_cache.NewTagsCacheOption()), app.TagAggView)
}
