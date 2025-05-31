// Path: ./conf/conf_upload.go

package conf

type Upload struct {
	ImageSizeLimit     int      `yaml:"imageSizeLimit"`
	ValidImageSuffixes []string `yaml:"validImageSuffixes"`
	ImageDir           string   `yaml:"imageDir"`
}
