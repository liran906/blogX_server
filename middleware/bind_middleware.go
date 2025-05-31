// Path: ./middleware/bind_middleware.go

package mdw

import (
	"blogX_server/common/res"
	"bytes"
	"github.com/gin-gonic/gin"
	"io"
)

func BindJsonMiddleware[T any](c *gin.Context) {
	// 注意 c 阅后即焚的特性，所以读取出来，后面每次读取都要再重新写入 c
	byteData, err := c.GetRawData()
	if err != nil {
		res.FailWithError(err, c)
		c.Abort()
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(byteData)) // 写回 c

	var req T
	err = c.ShouldBindJSON(&req)
	if err != nil {
		res.FailWithMsg("JSON 参数绑定错误: "+err.Error(), c)
		c.Abort()
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(byteData)) // 写回 c
	c.Set("bindReq", req)
}

func BindQueryMiddleware[T any](c *gin.Context) {
	// 注意 c 阅后即焚的特性，所以读取出来，后面每次读取都要再重新写入 c
	byteData, err := c.GetRawData()
	if err != nil {
		res.FailWithError(err, c)
		c.Abort()
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(byteData)) // 写回 c

	var req T
	err = c.ShouldBindQuery(&req)
	if err != nil {
		res.FailWithMsg("Query 参数绑定错误: "+err.Error(), c)
		c.Abort()
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(byteData)) // 写回 c
	c.Set("bindReq", req)
}

func BindUriMiddleware[T any](c *gin.Context) {
	// 注意 c 阅后即焚的特性，所以读取出来，后面每次读取都要再重新写入 c
	byteData, err := c.GetRawData()
	if err != nil {
		res.FailWithError(err, c)
		c.Abort()
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(byteData)) // 写回 c

	var req T
	err = c.ShouldBindUri(&req)
	if err != nil {
		res.FailWithMsg("URI 参数绑定错误: "+err.Error(), c)
		c.Abort()
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(byteData)) // 写回 c
	c.Set("bindReq", req)
}
