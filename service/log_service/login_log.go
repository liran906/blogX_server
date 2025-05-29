// Path: ./blogX_server/service/log_service/login_log.go

package log_service

import (
	"blogX_server/core"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"github.com/gin-gonic/gin"
)

// NewLoginSuccess 记录一条成功登录的登录日志
func NewLoginSuccess(user models.UserModel, loginType enum.LoginType, c *gin.Context) {
	//这里有问题啊，我就是因为没有 token 才需要登录
	/*
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
	*/

	// 入库
	ip := c.ClientIP()
	location, _ := core.GetLocationFromIP(ip)
	global.DB.Create(&models.LogModel{
		LogType:     enum.LoginLogType,
		Title:       "登录成功",
		Content:     "",
		UserID:      user.ID,
		IP:          ip,
		IPLocation:  location,
		LoginStatus: true,
		Username:    user.Username,
		Password:    "", // 成功登录不记录
		LoginType:   loginType,
		UA:          c.Request.UserAgent(),
	})
}

func NewLoginFail(loginType enum.LoginType, errMsg string, username string, pwd string, c *gin.Context) {
	ip := c.ClientIP()
	location, _ := core.GetLocationFromIP(ip)
	// 入库
	global.DB.Create(&models.LogModel{
		LogType:     enum.LoginLogType,
		Title:       "登录失败",
		Content:     errMsg,
		IP:          ip,
		IPLocation:  location,
		LoginStatus: false,
		Username:    username,
		Password:    pwd,
		LoginType:   loginType,
		UA:          c.Request.UserAgent(),
	})
}
