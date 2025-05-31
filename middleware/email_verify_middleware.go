// Path: ./middleware/email_verify_middleware.go

package mdw

import (
	"blogX_server/common/res"
	"blogX_server/global"
	"blogX_server/utils/email"
	"bytes"
	"github.com/gin-gonic/gin"
	"io"
)

type EmailVerifyMiddlewareRequest struct {
	EmailID   string `json:"emailID" binding:"required"`
	EmailCode string `json:"emailCode" binding:"required"`
}

func EmailVerifyMiddleware(c *gin.Context) {
	// 注意 c 阅后即焚的特性，所以读取出来，后面每次读取都要再重新写入 c
	byteData, err := c.GetRawData()
	if err != nil {
		res.FailWithError(err, c)
		c.Abort()
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(byteData)) // 写回 c

	var req EmailVerifyMiddlewareRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		res.FailWithError(err, c)
		c.Abort()
		return
	}

	// 校验验证码
	emailAddr, msg, uid, ok := email.IsValidEmailCode(req.EmailID, req.EmailCode)
	if !ok {
		res.FailWithMsg(msg, c)
		c.Abort()
		return
	}

	c.Set("email", emailAddr)
	c.Set("emailID", req.EmailID)
	c.Set("userIDFromEmailVerify", uid)
	c.Request.Body = io.NopCloser(bytes.NewReader(byteData)) // 写回 c
}

func EmailRegisterMiddleware(c *gin.Context) {
	if !global.Config.Site.Login.EmailRegister {
		res.FailWithMsg("站点未启用邮箱注册", c)
		c.Abort()
		return
	}
}
