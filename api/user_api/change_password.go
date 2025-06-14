// Path: ./api/user_api/change_password.go

package user_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/service/log_service"
	"blogX_server/utils/jwts"
	"blogX_server/utils/pwd"
	"blogX_server/utils/user"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"time"
)

type ChangePasswordReq struct {
	OldPassword string `json:"oldPassword"`
	Password    string `json:"password"`
}

func (UserApi) ChangePasswordView(c *gin.Context) {
	req := c.MustGet("bindReq").(ChangePasswordReq)

	if req.OldPassword == req.Password {
		res.FailWithMsg("修改后的密码不能相同", c)
		return
	}

	// 新密码强度校验
	if !user.IsValidPassword(req.Password) {
		res.FailWithMsg("密码不符合要求", c)
		return
	}

	// 获取身份
	claims, ok := jwts.GetClaimsFromRequest(c)
	if !ok {
		res.FailWithMsg("登录信息获取失败", c)
		return
	}

	// 读库
	u, err := claims.GetUserFromClaims()
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	// 当前密码为空
	if u.Password == "" {
		res.FailWithMsg("不符合修改密码条件", c)
		return
	}

	// 校验旧密码
	if !pwd.CompareHashAndPassword(u.Password, req.OldPassword) {
		res.FailWithMsg("当前密码输入错误", c)
		return
	}

	// 新密码加盐哈希
	hashPwd, err := pwd.GenerateFromPassword(req.Password)
	if err != nil {
		res.FailWithMsg("密码设置错误: "+err.Error(), c)
		return
	}

	// 入库
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

	// 日志
	log := log_service.GetActionLog(c)
	log.ShowRequestHeader()
	log.ShowResponseHeader()
	log.ShowResponse()
	log.SetTitle("更改密码")

	res.SuccessWithMsg("密码更新成功", c)
}
