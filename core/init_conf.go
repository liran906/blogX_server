package core

import (
	"blogX_server/flags"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type System struct {
	IP   string `yaml:"ip"`
	Port int    `yaml:"port"`
}

type Config struct {
	System System `yaml:"system"`
}

// ReadConf 读取 settings.yaml 设置文件
func ReadConf() {
	byteData, err := os.ReadFile(flags.FlagOptions.File)
	if err != nil {
		panic(err)
	}

	var config Config

	err = yaml.Unmarshal(byteData, &config)
	if err != nil {
		panic(fmt.Sprintln("yaml unmarshal err: ", err))
	}

	fmt.Printf("configuration of: %s success!\n", flags.FlagOptions.File)
}
