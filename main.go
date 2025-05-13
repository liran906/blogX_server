package main

import (
	"blogX_server/core"
	"blogX_server/flags"
)

func main() {
	flags.Parse()   // 解析命令行
	core.ReadConf() // 读取配置文件
}
