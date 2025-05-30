// Path: ./blogX_server/middleware/auth_middelware.go

package mdw

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models/enum"
	"blogX_server/service/redis_service/redis_jwt"
	"blogX_server/utils/jwts"
	"fmt"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(c *gin.Context) {
	claims, ok := getValidClaims(c)
	if !ok {
		c.Abort()
		return
	}
	c.Set("claims", claims)
}

func AdminMiddleware(c *gin.Context) {
	claims, ok := getValidClaims(c)
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

// getValidClaims extracts and validates JWT claims from the request, returning them if valid or responding with failure on error.
func getValidClaims(c *gin.Context) (claims *jwts.MyClaims, ok bool) {
	claims, err := jwts.ParseTokenFromRequest(c)
	if err != nil {
		res.FailWithError(err, c)
		return
	}
	if isExpiredToken(claims) {
		res.FailWithMsg("登录已过期", c)
		return
	}
	blockType, isBlocked := redis_jwt.IsBlockedJWTTokenByGin(c)
	if isBlocked {
		res.FailWithMsg(fmt.Sprintf("token 无效: %s", blockType.Msg()), c)
		return
	}
	return claims, true
}

// isExpiredToken checks if a token is expired by comparing its issued time with the password update timestamp in Redis.
func isExpiredToken(claims *jwts.MyClaims) bool {
	key := fmt.Sprintf("%dpassword_update", claims.UserID)
	// 先从 Redis 获取
	updateUnix, err := global.Redis.Get(key).Int64()
	if err == nil { // 找到了
		if updateUnix > claims.IssuedAt {
			return true
		}
	}
	return false
}
