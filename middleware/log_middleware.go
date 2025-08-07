// Path: ./middleware/log_middleware.go

package mdw

import (
	"blogX_server/service/log_service"
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

	// 把 log 对象存入 context 的 log 字段中，
	// 后续可以通过 log_service.GetActionLog() 方法，查询 log 字段判断是否为第一次创建
	// 以免在视图中重新创建一个 c，并重复入库
	c.Set("log", log)

	// 如果是 SSE 请求，保持原始的 Content-Type
	//if c.Request.URL.Path == "/api/ai_search" {
	//	c.Next()
	//	log.MiddlewareSave()
	//	return
	//}

	resWriter := &ResponseWriter{
		ResponseWriter: c.Writer,
		Head:           make(http.Header),
	}

	// 对于 SSE 响应，必须设置 ResponseWriter 的 Content-Type 为 text/event-stream
	// 如果不在这里处理，所有的 SSE 响应都无法达到效果（因为被日志截获覆盖了）
	// 这里只需要设置原始 ResponseWriter 的 header，因为：
	// 1. 我们自定义的 ResponseWriter 结构体会继承原始的 ResponseWriter
	// 2. 日志中间件会在所有 handler 执行前被调用
	// 3. SSE 需要在发送第一个事件前就设置好 header
	// 4. 如果不在中间件中设置，后续的 handler 中设置可能会被其他中间件覆盖

	// 定义需要使用 SSE (Server-Sent Events) 的路由列表
	streamURL := map[string]struct{}{
		"/api/ai_search": {}, // AI 搜索接口使用 SSE
	}

	// 获取当前请求的路径
	reqURL := c.Request.URL.Path

	// 检查当前请求是否需要 SSE
	// ok 为 true 表示该路由在 streamURL 中存在
	if _, ok := streamURL[reqURL]; ok {
		// 设置原始 ResponseWriter 的 header
		resWriter.ResponseWriter.Header().Set("Content-Type", "text/event-stream")
	}

	c.Writer = resWriter
	c.Next()

	// 响应中间件
	log.SetResponse(resWriter.Body)
	log.SetResponseHeader(resWriter.Head)

	log.MiddlewareSave()

	// 新增：AccessLog保存（每个请求都会记录）
	accessLog := log_service.NewAccessLog(c)
	accessLog.SetResponse(resWriter.Body)
	accessLog.Save()
}
