// Path: ./blogX_server/router/log_router.go

package router

import (
	"blogX_server/api"
	"github.com/gin-gonic/gin"
)

func LogRouter(r *gin.RouterGroup) {
	app := api.App.LogApi
	r.GET("/logs", app.LogListView)
	r.GET("/logs/:id", app.LogReadView)
}
