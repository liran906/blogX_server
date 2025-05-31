// Path: ./conf/conf_cloud.go

package conf

type Cloud struct {
	QNY QNY `yaml:"qny" json:"qny"`
}

type QNY struct {
	Enable    bool   `yaml:"enable" json:"enable"`       // 开启后默认会传到云
	LocalSave bool   `yaml:"localSave" json:"localSave"` // 开启云存储后，是否存到本地
	AccessKey string `yaml:"accessKey" json:"accessKey"`
	SecretKey string `yaml:"secretKey" json:"secretKey"`
	Bucket    string `yaml:"bucket" json:"bucket"`
	Uri       string `yaml:"uri" json:"uri"`
	Region    string `yaml:"region" json:"region"`
	Prefix    string `yaml:"prefix" json:"prefix"`
	Size      int    `yaml:"size" json:"size"` // 大小限制 单位mb
}
