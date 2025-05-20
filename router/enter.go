// Path: ./router/enter.go

package router

import (
	"blogX_server/global"
	"github.com/gin-gonic/gin"
	"log"
)

func Run() {
	gin.SetMode(global.Config.System.GinMode) // 设置 gin 模式，对应 settings.yaml 中的 gin_mode
	r := gin.Default()

	r.Static("/uploads", "uploads") // 配置静态路由访问上传文件

	nr := r.Group("/api")

	//nr.Use(middleware.LogMiddleware)
	SiteRouter(nr)
	LogRouter(nr)

	addr := global.Config.System.Addr()
	err := r.Run(addr)
	if err != nil {
		log.Fatalln(err)
	}
}
