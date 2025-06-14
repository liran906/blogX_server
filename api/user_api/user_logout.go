// Path: ./api/user_api/user_logout.go

package user_api

import (
	"blogX_server/common/res"
	"blogX_server/service/redis_service/redis_jwt"
	"blogX_server/utils/jwts"
	"github.com/gin-gonic/gin"
)

func (UserApi) UserLogoutView(c *gin.Context) {
	token, err := jwts.GetTokenFromRequest(c)
	if err != nil {
		res.Fail(err, "token获取失败", c)
		return
	}

	redis_jwt.BlockJWTToken(token, redis_jwt.UserBlockType)
	res.SuccessWithMsg("注销成功", c)
}
