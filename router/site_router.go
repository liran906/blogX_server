// Path: ./router/site_router.go

package router

import (
	"blogX_server/api"
	"github.com/gin-gonic/gin"
)

// SiteRouter 处理 api 路由分组
func SiteRouter(r *gin.RouterGroup) {
	app := api.App.SiteApi
	r.GET("/site", app.SiteInfoView) // 分别绑定到各个视图
}
