package site_api

import "github.com/gin-gonic/gin"

type SiteApi struct{}

// 每个路由绑定到一个视图（View），也就是对应一个页面

func (s *SiteApi) SiteInfoView(c *gin.Context) {
	// TBD
	c.JSON(200, gin.H{"message": "test: 站点信息"})
	return
}
