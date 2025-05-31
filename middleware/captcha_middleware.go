// Path: ./middleware/captcha_middleware.go

package mdw

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"bytes"
	"errors"
	"github.com/gin-gonic/gin"
	"io"
)

type CaptchaMiddlewareRequest struct {
	CaptchaCode string `json:"captchaCode" binding:"required"`
	CaptchaID   string `json:"captchaID" binding:"required"`
}

func CaptchaMiddleware(c *gin.Context) {
	if !global.Config.Site.Login.Captcha {
		return
	}

	// 注意 c 阅后即焚的特性，所以读取出来，后面每次读取都要再重新写入 c
	byteData, err := c.GetRawData()
	if err != nil {
		res.FailWithError(err, c)
		c.Abort()
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(byteData)) // 写回 c

	var captcha CaptchaMiddlewareRequest
	err = c.ShouldBindJSON(&captcha)
	if err != nil {
		res.FailWithError(errors.New("验证码缺失\n"+err.Error()), c)
		c.Abort()
		return
	}

	if !global.CaptchaStore.Verify(captcha.CaptchaID, captcha.CaptchaCode, true) {
		res.FailWithMsg("验证码错误", c)
		c.Abort()
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(byteData)) // 写回 c
}
