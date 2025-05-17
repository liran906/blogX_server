// Path: ./middleware/log_middleware.go

package middleware

import (
	"blogX_server/service/log_service"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

// ResponseWriter 由于 gin 的 RW 没有提供 read 接口，而我们想要读取返回的内容以便写入日志
// 所以我们自己实现一个 RW，继承自 gin.ResponseWriter
type ResponseWriter struct {
	gin.ResponseWriter
	Body []byte // 增加一个字段存储 Body
	Head http.Header
}

func (w *ResponseWriter) Write(b []byte) (int, error) {
	// write 方法中加入一步：写入 Body
	w.Body = append(w.Body, b...)
	// 然后继续调用原来的方法
	return w.ResponseWriter.Write(b)
}

func (w *ResponseWriter) Header() http.Header {
	return w.Head
}

func LogMiddleware(c *gin.Context) {
	// 请求中间件
	// 创建日志对象
	log := log_service.NewActionLogByGin(c)

	log.SetRequest(c)
	// 吧 log 对象存入 context 的 log 字段中， 后续可以通过查询 log 字段判断是否为第一次创建
	// 以免在视图中重新创建一个 c，并重复入库
	c.Set("log", log)

	res := &ResponseWriter{
		ResponseWriter: c.Writer,
		Head:           make(http.Header),
	}
	c.Writer = res
	c.Next()

	// 响应中间件
	fmt.Println("test2: ", res.Head)
	log.SetResponse(res.Body)
	log.SetResponseHeader(res.Head)
	log.Save()
}
