package log_service

import (
	"blogX_server/core"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"fmt"
	"github.com/gin-gonic/gin"
)

// NewLoginSuccess 记录一条成功登录的登录日志
func NewLoginSuccess(c *gin.Context, loginType enum.LoginType) {
	ip := c.ClientIP()
	address, _ := core.GetAddress(ip)

	token := c.GetHeader("token")

	// TBD 等讲了 JWT
	fmt.Println(token)
	userID := uint(1)
	username := ""

	// 入库
	global.DB.Create(&models.LogModel{
		LogType:     enum.LoginLogType,
		Title:       "登录成功",
		Content:     "",
		UserID:      userID,
		IP:          ip,
		Address:     address,
		LoginStatus: true,
		Username:    username,
		Password:    "", // 成功登录不记录
		LoginType:   loginType,
	})
}

func NewLoginFail(c *gin.Context, loginType enum.LoginType, errMsg string, username string, pwd string) {
	ip := c.ClientIP()
	address, _ := core.GetAddress(ip)

	// 入库
	global.DB.Create(&models.LogModel{
		LogType:     enum.LoginLogType,
		Title:       "登录失败",
		Content:     errMsg,
		IP:          ip,
		Address:     address,
		LoginStatus: false,
		Username:    username,
		Password:    pwd,
		LoginType:   loginType,
	})
}
