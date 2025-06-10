// Path: ./router/global_notification_router.go

package router

import (
	"blogX_server/api"
	"blogX_server/api/global_notification_api"
	mdw "blogX_server/middleware"
	"blogX_server/models"
	"github.com/gin-gonic/gin"
)

func GlobalNotificationRouter(rg *gin.RouterGroup) {
	app := api.App.GlobalNotificationApi

	rg.POST("global_notification", mdw.BindJsonMiddleware[global_notification_api.GNCreateReq], mdw.AdminMiddleware, app.GNCreateView)
	rg.GET("global_notification", mdw.BindQueryMiddleware[global_notification_api.GNListReq], mdw.AuthMiddleware, app.GNListView)
	rg.DELETE("global_notification", mdw.BindJsonMiddleware[models.IDListRequest], mdw.AdminMiddleware, app.GNRemoveView)
	rg.PUT("global_notification/read", mdw.BindJsonMiddleware[models.IDListRequest], mdw.AuthMiddleware, app.GNUserReadView)
	rg.PUT("global_notification/delete", mdw.BindJsonMiddleware[models.IDListRequest], mdw.AuthMiddleware, app.GNUserDeleteView)
}
