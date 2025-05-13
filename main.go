package main

import (
	"blogX_server/core"
	"blogX_server/flags"
	"blogX_server/global"
)

func main() {
	flags.Parse()                   // 解析命令行
	global.Config = core.ReadConf() // 读取配置文件
	core.InitLogrus()               // 初始化日志文件
	global.DB = core.InitDB()       // 链接数据库
}
