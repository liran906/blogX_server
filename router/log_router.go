// Path: ./router/log_router.go

package router

import (
	"blogX_server/api"
	"blogX_server/middleware"
	"github.com/gin-gonic/gin"
)

func LogRouter(rg *gin.RouterGroup) {
	// 绑定中间件，注意不能直接在传入的指针上使用，否则其他视图都会被绑定
	r := rg.Group("").Use(mdw.AdminMiddleware)

	// app 指向全局变量 App 的 LogApi 字段（LogApi 结构体，有对应方法）
	app := api.App.LogApi

	// 具体路由
	r.GET("/logs", app.LogListView)
	r.GET("/logs/:id", app.LogReadView)
	r.DELETE("/logs", app.LogRemoveView)
}
