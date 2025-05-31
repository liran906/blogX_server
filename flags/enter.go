// Package flags 提供命令行参数的解析和处理功能
// 这个包在整个项目中负责：
// 1. 参数定义和管理：统一管理所有的命令行参数
// 2. 配置灵活性：允许用户通过命令行选项自定义程序行为
// 3. 全局访问：其他包可以方便地访问解析后的参数值
// 4. 标准化处理：使用 Go 标准库的 flag 包提供统一的参数解析和验证机制
package flags

import (
	"flag"
	"os"
)

// Options 结构体定义了所有支持的命令行选项
// 这个结构体集中管理所有的命令行参数，使得参数管理更加规范和统一
type Options struct {
	// File 配置文件的路径
	// 可以通过 -f 参数指定，默认值为 "settings.yaml"
	// 使用示例：./program -f custom-config.yaml
	File string

	// DB 是否需要执行数据库迁移
	// 可以通过 -db 参数控制，默认为 false
	// 使用示例：./program -db
	DB bool

	// Version 是否显示版本信息
	// 可以通过 -v 参数控制，默认为 false
	// 使用示例：./program -v
	Version bool

	ES bool

	// Type 针对哪个类型进行操作
	// 可以通过 -t 参数控制
	Type string

	// Sub 针对哪个子类进行操作
	// 可以通过 -s 参数控制
	Sub string
}

// FlagOptions 是一个全局变量，用于存储解析后的命令行参数值
// 其他包可以通过这个变量访问用户通过命令行指定的配置
var FlagOptions = new(Options)

// Parse 初始化并解析命令行参数
// 该函数完成以下工作：
// 1. 定义所有支持的命令行参数及其默认值
// 2. 解析命令行输入
// 3. 将解析结果存储在 FlagOptions 中供程序其他部分使用
//
// 使用示例：
// 基本使用：./program
// 指定配置：./program -f custom-config.yaml
// 数据库迁移：./program -db
// 查看版本：./program -v
// 组合使用：./program -f custom-config.yaml -db
// 命令行创建用户：./program -t user -s create （可用于远程部署后创建一个管理员）
func Parse() {
	// 定义 -f 参数，用于指定配置文件路径
	// 当用户未指定时，默认使用 "settings.yaml" 作为配置文件
	flag.StringVar(&FlagOptions.File, "f", "settings.yaml", "config file")

	// 定义 -db 参数，用于控制是否执行数据库迁移
	// 这是一个布尔类型参数，默认为 false
	flag.BoolVar(&FlagOptions.DB, "db", false, "database migration")

	// 建立索引，如果之前有就把之前的删除（导出）重新建立（再导入之前的数据）
	flag.BoolVar(&FlagOptions.ES, "es", false, "ES init index")

	// 定义 -v 参数，用于控制是否显示版本信息
	// 这是一个布尔类型参数，默认为 false
	flag.BoolVar(&FlagOptions.Version, "v", false, "show version")

	flag.StringVar(&FlagOptions.Type, "t", "", "type")
	flag.StringVar(&FlagOptions.Sub, "s", "", "subtype")

	// 解析命令行参数
	// 这个调用会处理所有通过命令行传入的参数
	// 并将结果存储在相应的变量中
	flag.Parse()
}

// Run 调用函数，做数据库迁移
func Run() {
	if FlagOptions.DB {
		FlagDB()
		os.Exit(0)
	}

	if FlagOptions.ES {
		ESInitIndex()
		os.Exit(0)
	}

	switch FlagOptions.Type {
	case "user":
		u := FlagUser{}
		switch FlagOptions.Sub {
		case "create":
			u.Create()
			os.Exit(0)
		}
	}
}
