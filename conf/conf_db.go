package conf

import "fmt"

type DB struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	DB       string `yaml:"db"`
	Debug    bool   `yaml:"debug"`  // 打印全部日志
	Source   string `yaml:"source"` // 数据库源 mysql pgsql
}

func (d DB) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=true&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.DB,
	)
}

func (d DB) IsEmpty() bool {
	return d.User == "" && d.Password == "" && d.Host == "" && d.Port == 0
}
