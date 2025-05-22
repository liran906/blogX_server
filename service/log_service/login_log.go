// Path: ./service/log_service/login_log.go

package log_service

import (
	"blogX_server/common/res"
	"blogX_server/core"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/utils/jwts"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// NewLoginSuccess 记录一条成功登录的登录日志
func NewLoginSuccess(c *gin.Context, loginType enum.LoginType) {
	ip := c.ClientIP()
	address, _ := core.GetAddress(ip)

	// 从 token 中读取 uid
	var userID uint
	var username string
	claim, err := jwts.ParseTokenFromGin(c)
	if claim != nil && err == nil {
		userID, username = claim.UserID, claim.Username
	} else {
		// 这里按照教程没有，但我觉得应该报错+终止函数 TBD
		logrus.Errorf("failed to parse token: %v\n", err)
		res.FailWithError(err, c)
		return
	}

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
