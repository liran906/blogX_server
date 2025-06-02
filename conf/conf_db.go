// Path: ./conf/conf_db.go

package conf

import "fmt"

type DB struct {
	Name     string `yaml:"name"`     // db 的名字，比如 master slave 等
	User     string `yaml:"user"`     // db 登录用户名
	Password string `yaml:"password"` // db 登录密码
	Host     string `yaml:"host"`     // db ip 地址
	Port     int    `yaml:"port"`     // db 端口
	DBname   string `yaml:"dbname"`   // 哪个 database
	Debug    bool   `yaml:"debug"`    // 是否打印全部日志
	Source   string `yaml:"source"`   // 数据库源 mysql pgsql
}

func (d DB) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=true&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.DBname,
	)
}

func (d DB) IsEmpty() bool {
	return d.User == "" && d.Password == "" && d.Host == "" && d.Port == 0
}

func (d DB) GetAddr() string {
	return fmt.Sprintf("%s:%d", d.Host, d.Port)
}
