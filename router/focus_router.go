// Path: ./router/focus_router.go

package router

import (
	"blogX_server/api"
	"blogX_server/api/focus_api"
	"blogX_server/middleware"
	"github.com/gin-gonic/gin"
)

func FocusRouter(r *gin.RouterGroup) {
	app := api.App.FocusApi
	r.POST("focus", mdw.AuthMiddleware, mdw.BindJsonMiddleware[focus_api.FocusUserRequest], app.FocusUserView)
	r.GET("focus/my_focus", mdw.BindQueryMiddleware[focus_api.FocusUserListRequest], app.FocusUserListView)
	r.GET("focus/my_fans", mdw.BindQueryMiddleware[focus_api.FocusUserListRequest], app.FansUserListView)
	r.DELETE("focus", mdw.AuthMiddleware, mdw.BindJsonMiddleware[focus_api.FocusUserRequest], app.UnFocusUserView)

}
