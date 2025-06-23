// Path: ./api/enter.go

package api

import (
	"blogX_server/api/ai_api"
	"blogX_server/api/article_api"
	"blogX_server/api/banner_api"
	"blogX_server/api/captcha_api"
	"blogX_server/api/comment_api"
	"blogX_server/api/data_api"
	"blogX_server/api/focus_api"
	"blogX_server/api/global_notification_api"
	"blogX_server/api/image_api"
	"blogX_server/api/log_api"
	"blogX_server/api/mytest_api"
	"blogX_server/api/notify_api"
	"blogX_server/api/search_api"
	"blogX_server/api/site_api"
	"blogX_server/api/user_api"
)

type Api struct {
	SiteApi               site_api.SiteApi
	LogApi                log_api.LogApi
	ImageApi              image_api.ImageApi
	BannerApi             banner_api.BannerApi
	CaptchaApi            captcha_api.CaptchaApi
	UserApi               user_api.UserApi
	ArticleApi            article_api.ArticleApi
	CommentApi            comment_api.CommentApi
	NotifyApi             notify_api.NotifyApi
	GlobalNotificationApi global_notification_api.GlobalNotificationApi
	SearchApi             search_api.SearchApi
	AiApi                 ai_api.AiApi
	DataApi               data_api.DataApi
	FocusApi              focus_api.FocusApi

	MyTestApi mytest_api.MyTestApi // 测试用
}

var App = new(Api)
