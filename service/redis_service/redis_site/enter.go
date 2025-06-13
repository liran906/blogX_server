// Path: ./service/redis_service/redis_site/enter.go

package redis_site

import (
	"blogX_server/global"
	"time"
)

const flowKey = "blogx_site_flow"
const clickKey = "blogx_site_click"

// 获取当天的日期字符串
func getField() string {
	return time.Now().Format("2006-01-02")
}

// IncreaseFlow 站点流量+1，在站点信息接口调用（入站就会请求站点信息）
func IncreaseFlow() {
	global.Redis.HIncrBy(flowKey, getField(), 1)
}

func GetFlow(field string) int {
	flowCount, err := global.Redis.HGet(flowKey, field).Int()
	if err != nil {
		return 0
	}
	return flowCount
}

func IncreaseClick() {
	global.Redis.HIncrBy(clickKey, getField(), 1)
}

func GetClick(field string) int {
	clickCount, err := global.Redis.HGet(clickKey, field).Int()
	if err != nil {
		return 0
	}
	return clickCount
}

func ClearAllByDate(field string) {
	global.Redis.HDel(flowKey, field)
	global.Redis.HDel(clickKey, field)
}
