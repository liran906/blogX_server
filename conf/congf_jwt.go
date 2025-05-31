// Path: ./conf/congf_jwt.go

package conf

type Jwt struct {
	Expire int    `yaml:"expire"` // 单位：小时
	Secret string `yaml:"secret"`
	Issuer string `yaml:"issuer"`
}
