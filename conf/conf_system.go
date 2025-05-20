// Path: ./blogX_server/conf/conf_system.go

package conf

import "fmt"

// System 定义了系统配置的结构体
type System struct {
	IP      string `yaml:"ip"`       // 服务器监听的IP地址
	Port    int    `yaml:"port"`     // 服务器监听的端口号
	Env     string `yaml:"env"`      // 环境：dev/prod/test...
	GinMode string `yaml:"gin_mode"` // gin的模式：debug/release/test
}

func (s System) Addr() string {
	return fmt.Sprintf("%s:%d", s.IP, s.Port)
}
