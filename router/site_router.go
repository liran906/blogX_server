// Path: ./router/site_router.go

package router

import (
	"blogX_server/api"
	"blogX_server/middleware"
	"github.com/gin-gonic/gin"
)

// SiteRouter 处理 api 路由分组
func SiteRouter(rg *gin.RouterGroup) {
	// app 指向全局变量 App 的 SiteApi 字段（SiteApi 结构体，有对应方法）
	app := api.App.SiteApi

	// 下面通过 app（SiteApi）的方法，将对应视图分别绑定到路由
	rg.GET("/site/qq_url", app.SiteInfoQQView)
	rg.GET("/site/:name", app.SiteInfoView)
	rg.PUT("/site/:name", mdw.AdminMiddleware, app.SiteUpdateView)
}
