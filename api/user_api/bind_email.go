// Path: ./api/user_api/bind_email.go

package user_api

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/service/log_service"
	"blogX_server/utils/jwts"
	"fmt"
	"github.com/gin-gonic/gin"
)

func (UserApi) BindEmailView(c *gin.Context) {
	// 从中间件存储的 email 字段中拿取 email 地址
	emailAddr, ok := c.Get("email")
	if !ok {
		res.FailWithMsg("系统读取邮箱地址错误", c)
		return
	}

	// 取个 uid
	claims, ok := jwts.GetClaimsFromRequest(c)
	if !ok {
		res.FailWithMsg("请登录", c)
		return
	}
	// 比对申请绑定邮箱时的 uid，边界情况之一
	uid := c.MustGet("userIDFromEmailVerify").(uint) // 中间件绑定的
	fmt.Println(uid, claims.UserID)
	if uid != claims.UserID {
		res.FailWithMsg("用户错误", c)
		return
	}

	// 写库
	err := global.DB.Model(&models.UserModel{}).Where("ID = ?", claims.UserID).Update("email", emailAddr).Error
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	// 日志
	log := log_service.GetActionLog(c)
	log.ShowAll()
	log.SetTitle("邮箱绑定")

	// redis 删除记录
	eid := c.MustGet("emailID").(string)
	global.Redis.Del(eid)

	res.SuccessWithMsg("邮箱绑定成功", c)
}
