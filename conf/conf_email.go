// Path: ./conf/conf_email.go

package conf

type Email struct {
	Domain     string `yaml:"domain" json:"domain"`
	Port       int    `yaml:"port" json:"port"`
	SendEmail  string `yaml:"sendEmail" json:"sendEmail"`
	AuthCode   string `yaml:"authCode" json:"authCode"` // 授权码
	Alias      string `yaml:"alias" json:"alias"`       // 发送别名
	SSL        bool   `yaml:"ssl" json:"ssl"`
	TLS        bool   `yaml:"tls" json:"tls"`
	CodeExpiry int    `yaml:"codeExpiry" json:"codeExpiry"` // 单位分钟
}
