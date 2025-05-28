// Path: ./blogX_server/middleware/register_verify_middleware.go

package middleware

import (
	"blogX_server/common/res"
	"blogX_server/utils/user"
	"bytes"
	"github.com/gin-gonic/gin"
	"io"
)

type RegisterVerifyMiddlewareRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func RegisterVerifyMiddleware(c *gin.Context) {
	// 注意 c 阅后即焚的特性，所以读取出来，后面每次读取都要再重新写入 c
	byteData, err := c.GetRawData()
	if err != nil {
		res.FailWithError(err, c)
		c.Abort()
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(byteData)) // 写回 c

	var req RegisterVerifyMiddlewareRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		res.FailWithError(err, c)
		c.Abort()
		return
	}

	// 判断是合法 username
	msg, ok := user.IsValidUsername(req.Username)
	if !ok {
		res.FailWithMsg(msg, c)
		c.Abort()
		return
	}

	// 判断 username 是否重复
	msg, ok = user.IsAvailableUsername(req.Username)
	if !ok {
		res.FailWithMsg(msg, c)
		c.Abort()
		return
	}

	// 判断密码强度
	if !user.IsValidPassword(req.Password) {
		res.SuccessWithMsg("密码不符合要求", c)
		c.Abort()
		return
	}

	c.Request.Body = io.NopCloser(bytes.NewReader(byteData)) // 写回 c
}
