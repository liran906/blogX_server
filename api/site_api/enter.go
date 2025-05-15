// Path: ./api/site_api/enter.go

package site_api

import (
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"github.com/gin-gonic/gin"
)

type SiteApi struct{}

// 每个路由绑定到一个视图（View），也就是对应一个页面

func (s *SiteApi) SiteInfoView(c *gin.Context) {
	// TBD
	log_service.NewLoginSuccess(c, enum.UsernamePasswordLoginType)
	log_service.NewLoginFail(c, enum.UsernamePasswordLoginType, "login fail", "un_test", "pw_test")
	c.JSON(200, gin.H{"message": "test: 站点信息"})
	return
}
