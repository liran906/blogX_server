// Path: ./conf/conf_filter.go

package conf

type Filter struct {
	// 黑名单
	InvalidUsername []string `yaml:"invalidUsername"`

	// 白名单
	ValidImageSuffix []string `yaml:"validImageSuffix"`
}
