// Path: ./router/enter.go

package router

import (
	"blogX_server/global"
	"blogX_server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"log"
)

func Run() {
	gin.SetMode(global.Config.System.GinMode) // 设置 gin 模式，对应 settings.yaml 中的 gin_mode
	r := gin.Default()

	r.Static("/uploads", "uploads") // 配置静态路由访问上传文件

	nr := r.Group("/api")

	// 全部记录日志
	nr.Use(mdw.LogMiddleware)

	// 具体路由
	LogRouter(nr)
	ImageRouter(nr)
	SiteRouter(nr)
	BannerRouter(nr)
	CaptchaRouter(nr)
	UserRouter(nr)
	ArticleRouter(nr)
	CommentRouter(nr)
	NotifyRouter(nr)
	GlobalNotificationRouter(nr)
	SearchRouter(nr)

	MytestRouter(nr) // 测试用

	addr := global.Config.System.Addr()
	logrus.Infof("Server running at %s", addr)
	err := r.Run(addr)
	if err != nil {
		log.Fatalln(err)
	}
}
