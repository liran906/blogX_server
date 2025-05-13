package conf

// Config 是整个应用的配置结构体
type Config struct {
	System System `yaml:"system"` // 系统配置部分
	Log    Log    `yaml:"log"`
	DB_r   DB     `yaml:"dbr"` // 读库，这里是读写分离，也可以不分
	DB_w   DB     `yaml:"dbw"` // 写库
}
