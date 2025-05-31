// Path: ./blogX_server/global/enter.go

package global

import (
	"blogX_server/conf"
	"github.com/go-redis/redis"
	"github.com/mojocn/base64Captcha"
	"gorm.io/gorm"
)

const Version = "10.0.1"

var (
	Config       *conf.Config
	DB           *gorm.DB
	DBMaster     *gorm.DB
	Redis        *redis.Client
	CaptchaStore = base64Captcha.DefaultMemStore
	Test         string
)
