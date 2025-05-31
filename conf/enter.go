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
	DB_r  DB    `yaml:"dbr"` // 读库，这里是读写分离，也可以不分
	DB_w  DB    `yaml:"dbw"` // 写库
	ES    ES    `yaml:"es"`

	// 站点设置
	Site   Site   `yaml:"site"`
	Ai     Ai     `yaml:"ai"`
	Cloud  Cloud  `yaml:"cloud"`
	QQ     QQ     `yaml:"qq"`
	Email  Email  `yaml:"email"`
	Upload Upload `yaml:"upload"`
}
