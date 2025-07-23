// Path: ./core/init_conf.go

// Package core 提供了博客服务器的核心功能
package core

import (
	"blogX_server/conf"
	"blogX_server/flags"
	"blogX_server/global"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// ReadConf 读取 settings.yaml 设置文件并解析配置
// 如果读取或解析过程中出现错误，将会触发panic
func ReadConf() (c *conf.Config) {
	// 从指定的配置文件路径读取内容
	logrus.Info("读取配置文件: ", flags.FlagOptions.File)
	byteData, err := os.ReadFile(flags.FlagOptions.File)
	if err != nil {
		panic(err)
	}

	// new 返回的就是指针
	c = new(conf.Config)

	// 将YAML格式的配置文件内容解析到config结构体中
	err = yaml.Unmarshal(byteData, c)
	if err != nil {
		panic(fmt.Sprintln("yaml unmarshal err: ", err))
	}

	// 打印配置文件读取成功的消息
	fmt.Printf("configuration of: %s success!\n", flags.FlagOptions.File)

	return
}

func SetConf() {
	byteData, err := yaml.Marshal(global.Config)
	if err != nil {
		logrus.Errorln("yaml marshal err: ", err)
		return
	}
	err = os.WriteFile("settings.yaml", byteData, 0666)
	if err != nil {
		logrus.Errorln("yaml write err: ", err)
		return
	}
	logrus.Info("settings.yaml write successful")
}
