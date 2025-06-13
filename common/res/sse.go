// Path: ./common/res/sse.go

package res

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
)

func SSESuccess(data any, c *gin.Context) {
	byteData, _ := json.Marshal(Response{Code: SuccessCode, Msg: "success", Data: data})
	c.SSEvent("", string(byteData))
	c.Writer.Flush()
}

func SSEFail(msg string, c *gin.Context) {
	byteData, _ := json.Marshal(Response{Code: FailServiceCode, Msg: msg, Data: struct{}{}})
	c.SSEvent("", string(byteData))
	c.Writer.Flush()
}
