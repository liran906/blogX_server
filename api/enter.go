// Path: ./blogX_server/api/enter.go

package api

import (
	"blogX_server/api/image_api"
	"blogX_server/api/log_api"
	"blogX_server/api/site_api"
)

type Api struct {
	SiteApi  site_api.SiteApi
	LogApi   log_api.LogApi
	ImageAip image_api.ImageApi
}

var App = new(Api)
