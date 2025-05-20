// Path: ./blogX_server/common/res/enter.go

package res

import "github.com/gin-gonic/gin"

type Code uint

const (
	SuccessCode        Code = 0
	FailValidationCode Code = 1001
	FailServiceCode    Code = 1002
)

func (c Code) ToString() string {
	switch c {
	case SuccessCode:
		return "Success"
	case FailValidationCode:
		return "Validation Failed"
	case FailServiceCode:
		return "Service Failed"
	}
	return ""
}

// 空值
var empty = map[string]any{}

type Response struct {
	Code Code   `json:"code"`
	Data any    `json:"data"`
	Msg  string `json:"msg"`
}

func (r Response) Json(c *gin.Context) {
	c.JSON(200, r)
}

func Success(data any, msg string, c *gin.Context) {
	Response{SuccessCode, data, msg}.Json(c)
}

func SuccessWithData(data any, c *gin.Context) {
	Response{SuccessCode, data, "Success"}.Json(c)
}

func SuccessWithMsg(msg string, c *gin.Context) {
	Response{SuccessCode, empty, msg}.Json(c)
}

func SuccessWithList(list any, count int, c *gin.Context) {
	Response{SuccessCode, map[string]any{
		"list":  list,
		"count": count,
	}, "Success"}.Json(c)
}

func FailWithMsg(msg string, c *gin.Context) {
	Response{FailValidationCode, empty, msg}.Json(c)
}

func FailWithData(data any, msg string, c *gin.Context) {
	Response{FailServiceCode, data, msg}.Json(c)
}

func FailWithCode(code Code, c *gin.Context) {
	Response{code, empty, code.ToString()}.Json(c)
}

func FailWithError(err error, c *gin.Context) {
	FailWithMsg(err.Error(), c)
}
