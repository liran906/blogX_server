// Path: ./blogX_server/conf/conf_ai.go

package conf

type Ai struct {
	Enable    bool   `yaml:"enable" json:"enable"`
	SecretKey string `yaml:"secretKey" json:"secretKey"`
	Nickname  string `yaml:"nickname" json:"nickname"`
	AvatarURL string `yaml:"avatarURL" json:"avatarURL"`
	Abstract  string `yaml:"abstract" json:"abstract"`
}
