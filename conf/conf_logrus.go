// Path: ./blogX_server/conf/conf_logrus.go

package conf

// Log 定义了日志配置的结构体
type Log struct {
	App string `yaml:"app"` // 应用程序名称，用于日志标识
	Dir string `yaml:"dir"` // 日志文件存储目录
}
