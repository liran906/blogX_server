// Path: ./blogX_server/api/user_api/change_password.go

package user_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/service/log_service"
	"blogX_server/utils/jwts"
	"blogX_server/utils/pwd"
	"blogX_server/utils/user"
	"github.com/gin-gonic/gin"
)

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	Password    string `json:"password"`
}

func (UserApi) ChangePasswordView(c *gin.Context) {
	var req ChangePasswordRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	// 新密码强度校验
	if !user.IsValidPassword(req.Password) {
		res.FailWithMsg("密码不符合要求", c)
		return
	}

	// 获取身份
	claims, ok := jwts.GetClaimsFromGin(c)
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
	err = global.DB.Model(&u).Update("password", hashPwd).Error
	if err != nil {
		res.FailWithError(err, c)
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
