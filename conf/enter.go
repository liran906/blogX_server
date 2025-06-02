// Path: ./conf/enter.go

package conf

// Config 是整个应用的配置结构体
type Config struct {
	System System `yaml:"system"` // 系统配置
	Jwt    Jwt    `yaml:"jwt"`    // jwt配置
	Log    Log    `yaml:"log"`    // 日志
	Filter Filter `yaml:"filter"` // 非法字段过滤

	// 数据库相关
	Redis Redis `yaml:"redis"`
	DB    []DB  `yaml:"db"` // 数据库连接列表
	ES    ES    `yaml:"es"`
	River River `yaml:"river"`

	// 站点设置
	Site   Site   `yaml:"site"`
	Ai     Ai     `yaml:"ai"`
	Cloud  Cloud  `yaml:"cloud"`
	QQ     QQ     `yaml:"qq"`
	Email  Email  `yaml:"email"`
	Upload Upload `yaml:"upload"`
}
