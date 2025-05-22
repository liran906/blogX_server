// Path: ./blogX_server/router/log_router.go

package router

import (
	"blogX_server/api"
	"blogX_server/middleware"
	"github.com/gin-gonic/gin"
)

func LogRouter(r *gin.RouterGroup) {
	// 绑定中间件
	r.Use(middleware.AdminMiddleware)

	// app 指向全局变量 App 的 LogApi 字段（LogApi 结构体，有对应方法）
	app := api.App.LogApi

	// 具体路由
	r.GET("/logs", app.LogListView)
	r.GET("/logs/:id", app.LogReadView)
	r.DELETE("/logs", app.LogRemoveView)
}
