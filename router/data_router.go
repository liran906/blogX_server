// Path: ./router/data_router.go

package router

import (
	"blogX_server/api"
	"blogX_server/api/data_api"
	mdw "blogX_server/middleware"
	"github.com/gin-gonic/gin"
)

func DataRouter(rg *gin.RouterGroup) {
	app := api.App.DataApi

	rg.GET("data", mdw.BindJsonMiddleware[data_api.SiteStatisticsReq], mdw.AdminMiddleware, app.SiteStatisticsView)
	rg.GET("data/week", mdw.BindQueryMiddleware[data_api.WeeklyGrowthDataReq], mdw.AdminMiddleware, app.WeeklyGrowthDataView)
}
