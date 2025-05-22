// Path: ./blogX_server/middleware/auth_middelware.go

package middleware

import (
	"blogX_server/common/res"
	"blogX_server/models/enum"
	"blogX_server/service/redis_service/redis_jwt"
	"blogX_server/utils/jwts"
	"fmt"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		c.Abort()
		return
	}
	c.Set("claims", claims)
}

func AdminMiddleware(c *gin.Context) {
	claims, ok := getClaims(c)
	if !ok {
		c.Abort()
		return
	}
	if claims.Role != enum.AdminRoleType {
		res.FailWithMsg("权限不足 (Not admin)", c)
		c.Abort()
		return
	}
	c.Set("claims", claims)
}

func getClaims(c *gin.Context) (claims *jwts.MyClaims, ok bool) {
	claims, err := jwts.ParseTokenFromGin(c)
	if err != nil {
		res.FailWithError(err, c)
		return
	}
	blockType, isBlocked := redis_jwt.IsBlockedJWTTokenByGin(c)
	if isBlocked {
		res.FailWithMsg(fmt.Sprintf("token is blocked: %s", blockType.Msg()), c)
		return
	}
	return claims, true
}
