// Path: ./blogX_server/api/user_api/reset_password.go

package user_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/service/log_service"
	"blogX_server/utils/pwd"
	"blogX_server/utils/user"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"time"
)

type ResetPasswordReq struct {
	Password string `json:"password" binding:"required"`
}

func (UserApi) ResetPasswordView(c *gin.Context) {
	req := c.MustGet("bindReq").(ResetPasswordReq)

	// 判断密码强度
	if !user.IsValidPassword(req.Password) {
		res.FailWithMsg("密码不符合要求", c)
		return
	}

	// 读库
	email := c.MustGet("email").(string)
	var u models.UserModel
	err := global.DB.Take(&u, "email = ?", email).Error
	if err != nil {
		res.FailWithMsg("读取邮箱错误 "+err.Error(), c)
		return
	}

	// 密码加盐并哈希
	hashPwd, err := pwd.GenerateFromPassword(req.Password)
	if err != nil {
		res.FailWithMsg("密码设置错误: "+err.Error(), c)
		return
	}

	// 写库
	err = global.DB.Model(&u).Updates(map[string]interface{}{
		"password":        hashPwd,
		"password_update": time.Now().Unix(),
	}).Error
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	// redis 缓存一个更新时间，以便验证之前签发的 token 为失效
	key := fmt.Sprintf("%dpassword_update", u.ID)
	err = global.Redis.Set(key, time.Now().Unix(), time.Duration(global.Config.Jwt.Expire)*time.Hour).Err()
	if err != nil {
		logrus.Error("缓存密码更新时间失败:", err)
		res.FailWithMsg("缓存密码更新时间失败: "+err.Error(), c)
		return
	}

	eid := c.MustGet("emailID").(string)
	global.Redis.Del(eid)

	// 日志
	log := log_service.GetActionLog(c)
	log.ShowRequestHeader()
	log.ShowResponseHeader()
	log.ShowResponse()
	log.SetTitle("邮箱重置密码")
	log.SetUID(u.ID)

	res.SuccessWithMsg("密码重置成功", c)
}
