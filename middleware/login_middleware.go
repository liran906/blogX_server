// Path: ./blogX_server/middleware/login_middleware.go

package middleware

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"github.com/gin-gonic/gin"
)

func UsernamePwdLoginMiddleware(c *gin.Context) {
	if !global.Config.Site.Login.UsernamePwdLogin {
		res.FailWithMsg("站点未开启用户名密码登录", c)
		c.Abort()
		return
	}
}
