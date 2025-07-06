// Path: ./middleware/cache_middleware.go

package mdw

import (
	"blogX_server/global"
	"blogX_server/service/redis_service/redis_cache"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/url"
)

type CacheResponseWriter struct {
	gin.ResponseWriter
	Body []byte
}

func (w *CacheResponseWriter) Write(data []byte) (int, error) {
	w.Body = append(w.Body, data...)
	return w.ResponseWriter.Write(data)
}

// CacheMiddleware 是一个通用的缓存中间件工厂函数，接受一个 CacheOption 参数用于配置。
// 它会尝试从 Redis 缓存中获取响应内容，如果命中则直接返回缓存结果，避免后续 handler 的执行。
// 如果未命中缓存，则会缓存 handler 的响应结果，用于后续请求复用。
func CacheMiddleware(option redis_cache.CacheOption) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ========== 请求阶段 ==========
		// 构建用于缓存键的查询参数字符串（按顺序拼接）
		var val string
		if option.IsUri {
			// 默认使用 "id" 作为 URI 参数名（如果未指定）
			paramKey := "id"
			if option.UriKey != "" {
				paramKey = option.UriKey
			}
			val = c.Param(paramKey)
		} else {
			values := url.Values{}
			for _, key := range option.Params {
				values.Add(key, c.Query(key))
			}
			val = values.Encode()
		}

		key := string(option.Prefix) + val

		// 尝试从 Redis 获取缓存数据
		val, err := global.Redis.Get(key).Result()
		fmt.Println(key, err)

		// 如果命中缓存，且未设置强制刷新（NoCache 为 nil 或 NoCache(c) 返回 false）
		if (err == nil && option.NoCache == nil) || (err == nil && option.NoCache(c) == false) {
			// 中止请求，阻止进入后续 handler
			c.Abort()

			//fmt.Println("!!!!!!!!!!!!!!!!!!!!!走缓存了")

			// 设置响应头为 JSON
			c.Header("Content-Type", "application/json; charset=utf-8")

			// 直接返回缓存内容
			c.Writer.Write([]byte(val))
			return
		}

		// 如果没有缓存没有命中，准备截获响应写入缓存，这里的逻辑同 log 的中间件
		// 初始化一个自定义的 ResponseWriter，用于拦截写入响应的数据
		w := &CacheResponseWriter{
			ResponseWriter: c.Writer, // 保留原始 Writer，后续仍要向客户端输出
		}
		// 替换 Gin 默认的响应 Writer
		c.Writer = w

		c.Next()
		// ========== 响应阶段 ==========
		// 拿到 handler 的响应体（已被拦截保存在 w.Body 中）
		body := string(w.Body)

		// 将响应内容写入 Redis 缓存，供后续相同请求使用
		redis_cache.CacheOpen(key, body, option.Expiry)
	}
}
