// Path: ./router/notify_router.go

package router

import (
	"blogX_server/api"
	"blogX_server/api/notify_api"
	mdw "blogX_server/middleware"
	"github.com/gin-gonic/gin"
)

func NotifyRouter(rg *gin.RouterGroup) {
	app := api.App.NotifyApi

	rg.GET("notify", mdw.BindQueryMiddleware[notify_api.NotifyListReq], mdw.AuthMiddleware, app.NotifyListView)
	rg.PUT("notify/read", mdw.BindJsonMiddleware[notify_api.NotifyReadReq], mdw.AuthMiddleware, app.NotifyReadView)
}
