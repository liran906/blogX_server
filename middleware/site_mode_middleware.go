// Path: ./middleware/site_mode_middleware.go

package mdw

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models/enum"
	"blogX_server/utils/jwts"
	"github.com/gin-gonic/gin"
)

// VerifySiteModeMiddleware 博客模式下，普通用户无法操作
func VerifySiteModeMiddleware(c *gin.Context) {
	if global.Config.Site.SiteInfo.Mode == 2 {
		claims := jwts.MustGetClaimsFromGin(c)
		if claims.Role != enum.AdminRoleType {
			res.FailWithMsg("当前站点为博客模式，无法进行此操作", c)
			c.Abort()
			return
		}
	}
}
