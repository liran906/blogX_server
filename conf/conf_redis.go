// Path: ./conf/conf_redis.go

package conf

type Redis struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	SyncTime string `yaml:"syncTime"` // 同步时间 eg. "0 0 2 * * *" 秒 分 小时 日 月 周
}
