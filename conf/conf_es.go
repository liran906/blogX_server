// Path: ./conf/conf_es.go

package conf

type ES struct {
	Addr     string `yaml:"addr"`
	IsHttps  bool   `yaml:"isHttps"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func (e ES) GetURL() string {
	if e.IsHttps {
		return "https://" + e.Addr
	}
	return "http://" + e.Addr
}
