// Path: ./blogX_server/router/site_router.go

package router

import (
	"blogX_server/api"
	"blogX_server/middleware"
	"github.com/gin-gonic/gin"
)

// SiteRouter 处理 api 路由分组
func SiteRouter(r *gin.RouterGroup) {
	// app 指向全局变量 App 的 SiteApi 字段（SiteApi 结构体，有对应方法）
	app := api.App.SiteApi

	// 下面通过 app（SiteApi）的方法，将对应视图分别绑定到路由
	r.GET("/site/qq_url", app.SiteInfoQQView)
	r.GET("/site/:name", app.SiteInfoView)
	r.PUT("/site/:name", middleware.AdminMiddleware, app.SiteUpdateView)
}
