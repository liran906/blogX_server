// Path: ./api/enter.go

package api

import "blogX_server/api/site_api"

type Api struct {
	SiteApi site_api.SiteApi
}

var App = new(Api)
