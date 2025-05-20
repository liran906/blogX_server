// Path: ./global/enter.go

package global

import (
	"blogX_server/conf"
	"gorm.io/gorm"
)

var (
	Config *conf.Config
	DB     *gorm.DB
)
