// Path: ./blogX_server/api/user_api/pwd_login.go

package user_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/service/email_service"
	"blogX_server/service/log_service"
	"blogX_server/utils/jwts"
	"blogX_server/utils/pwd"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

type PwdLoginRequest struct {
	Username string `json:"username" binding:"required"` // 可能是 username 也可能是邮箱地址
	Password string `json:"password" binding:"required"`
}

func (UserApi) PwdLoginView(c *gin.Context) {
	var req PwdLoginRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	var user models.UserModel
	var loginType enum.LoginType // 为了日志记录类型

	// 判断是邮箱还是用户名
	if email_service.IsValidEmail(req.Username) {
		loginType = enum.EmailPasswordLoginType
		err = global.DB.Take(&user, "email = ?", req.Username).Error
	} else {
		loginType = enum.UsernamePasswordLoginType
		err = global.DB.Take(&user, "username = ?", req.Username).Error
	}
	if err != nil || !pwd.CompareHashAndPassword(user.Password, req.Password) {
		msg := fmt.Sprintf("%s, 密码错误", loginType)
		if err != nil {
			msg = fmt.Sprintf("%s, 用户名错误 %s", loginType, err.Error())
		}
		log_service.NewLoginFail(loginType, msg, req.Username, req.Password, c)
		res.FailWithMsg("用户名或密码错误", c)
		return
	}

	// 颁发 token
	token, err := jwts.GenerateToken(jwts.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
	})
	if err != nil {
		res.FailWithMsg("token失败: "+err.Error(), c)
		return
	}

	// 登录成功
	global.DB.Model(&user).Updates(map[string]interface{}{
		"last_login_ip":   c.ClientIP(),
		"last_login_time": time.Now(),
	})

	// 登录日志
	log_service.NewLoginSuccess(user, loginType, c)

	// 返回 token 与成功信息
	res.Success(token, "登录成功", c)
}
