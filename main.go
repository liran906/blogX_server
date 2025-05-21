// Path: ./main.go

package main

import (
	"blogX_server/core"
	"blogX_server/flags"
	"blogX_server/global"
	"blogX_server/router"
)

func main() {
	flags.Parse()                   // 解析命令行
	global.Config = core.ReadConf() // 读取配置文件
	core.InitLogrus()               // 初始化日志文件
	global.DB = core.InitDB()       // 连接 mysql
	global.Redis = core.InitRedis() // 连接 redis
	flags.Run()                     // 数据库迁移
	router.Run()                    // 启动 web 服务
}
