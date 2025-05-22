// Path: ./blogX_server/conf/conf_cloud.go

package conf

type Cloud struct {
	Enable    bool   `yaml:"enable" json:"enable"`
	AccessKey string `yaml:"accessKey" json:"accessKey"`
	SecretKey string `yaml:"secretKey" json:"secretKey"`
	Bucket    string `yaml:"bucket" json:"bucket"`
	Uri       string `yaml:"uri" json:"uri"`
	Region    string `yaml:"region" json:"region"`
	Prefix    string `yaml:"prefix" json:"prefix"`
	Size      int    `yaml:"size" json:"size"`     // 大小限制 单位mb
	Expiry    int    `yaml:"expiry" json:"expiry"` // 过期时间 单位秒
}
