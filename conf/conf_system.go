package conf

// System 定义了系统配置的结构体
type System struct {
	IP   string `yaml:"ip"`   // 服务器监听的IP地址
	Port int    `yaml:"port"` // 服务器监听的端口号
}
