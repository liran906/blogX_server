// Path: ./router/enter.go

package router

import (
	"blogX_server/global"
	"github.com/gin-gonic/gin"
	"log"
)

func Run() {
	r := gin.Default()

	nr := r.Group("/api")
	SiteRouter(nr)

	addr := global.Config.System.Addr()
	err := r.Run(addr)
	if err != nil {
		log.Fatalln(err)
	}
}
