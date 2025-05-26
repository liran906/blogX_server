// Path: ./blogX_server/api/site_api/enter.go

package site_api

import (
	"blogX_server/common/res"
	"blogX_server/conf"
	"blogX_server/core"
	"blogX_server/global"
	"blogX_server/middleware"
	"blogX_server/models/enum"
	"blogX_server/service/log_service"
	"blogX_server/utils/jwts"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
)

type SiteApi struct{}

type SiteInfoRequest struct {
	Name string `uri:"name"`
}

// 每个路由绑定到一个视图（View），也就是对应一个页面

// SiteInfoView 查看站点配置
func (s *SiteApi) SiteInfoView(c *gin.Context) {
	var req SiteInfoRequest
	err := c.ShouldBindUri(&req)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	if req.Name == "site" {
		global.Config.Site.About.Version = global.Version // 不从配置文件读取，从 global cost中读取
		res.SuccessWithData(global.Config.Site, c)
		return
	}

	// 其余需要管理员权限，所以要先判断身份
	middleware.AdminMiddleware(c)
	_cla, ok := c.Get("claims")
	if !ok {
		return
	}

	cla, ok := _cla.(*jwts.MyClaims)
	if !ok || cla.Role != enum.AdminRoleType {
		return
	}

	var data any

	switch req.Name {
	case "ai":
		rep := global.Config.Ai
		rep.SecretKey = "******" // 保护敏感信息
		data = rep
	case "cloud":
		rep := global.Config.Cloud.QNY
		rep.SecretKey = "******" // 保护敏感信息
		data = rep
	case "qq":
		rep := global.Config.QQ
		rep.AppKey = "******" // 保护敏感信息
		data = rep
	case "email":
		rep := global.Config.Email
		rep.AuthCode = "******" // 保护敏感信息
		data = rep
	default:
		res.FailWithMsg("Unknown site name: "+req.Name, c)
		return
	}

	res.SuccessWithData(data, c)
}

// SiteInfoQQView qq登录
func (s *SiteApi) SiteInfoQQView(c *gin.Context) {
	res.SuccessWithData(global.Config.QQ.Url(), c)
}

type SiteUpdateRequest struct {
	Name string `json:"name" binding:"required"`
}

// SiteUpdateView 更新站点配置
func (s *SiteApi) SiteUpdateView(c *gin.Context) {
	// 写入日志
	log := log_service.GetActionLog(c)
	log.ShowAll()
	log.SetTitle("站点配置更新失败")

	var r SiteInfoRequest
	err := c.ShouldBindUri(&r)
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	var rep any
	switch r.Name {
	case "site":
		var data = global.Config.Site
		err = c.ShouldBindJSON(&data)
		rep = data
	case "ai":
		var data = global.Config.Ai
		err = c.ShouldBindJSON(&data)
		rep = data
	case "cloud":
		var data = global.Config.Cloud
		err = c.ShouldBindJSON(&data)
		rep = data
	case "email":
		var data = global.Config.Email
		err = c.ShouldBindJSON(&data)
		rep = data
	case "qq":
		var data = global.Config.QQ
		err = c.ShouldBindJSON(&data)
		rep = data
	default:
		res.FailWithMsg("Unknown site name: "+r.Name, c)
		return
	}
	if err != nil {
		res.FailWithError(err, c)
		return
	}

	switch s := rep.(type) {
	case conf.Site:
		err = UpdateSite(s)
		if err != nil {
			res.FailWithError(err, c)
			return
		}
		global.Config.Site = s
	case conf.Ai:
		if s.SecretKey == "******" {
			s.SecretKey = global.Config.Ai.SecretKey
		}
		global.Config.Ai = s
	case conf.Cloud:
		if s.QNY.SecretKey == "******" {
			s.QNY.SecretKey = global.Config.Cloud.QNY.SecretKey
		}
		global.Config.Cloud = s
	case conf.Email:
		if s.AuthCode == "******" {
			s.AuthCode = global.Config.Email.AuthCode
		}
		global.Config.Email = s
	case conf.QQ:
		if s.AppKey == "******" {
			s.AppKey = global.Config.QQ.AppKey
		}
		global.Config.QQ = s
	}

	// 改配置文件
	core.SetConf()

	log.SetTitle("站点配置更新成功")

	res.SuccessWithMsg("站点配置更新成功", c)
}

// UpdateSite 更新前端文件
func UpdateSite(site conf.Site) error {
	if site.Project.Icon == "" && site.Project.Title == "" && site.Project.WebPath == "" &&
		site.Seo.Description == "" && site.Seo.Keywords == "" {
		return nil
	}

	if site.Project.WebPath == "" {
		return errors.New("请配置前端地址")
	}

	file, err := os.Open(site.Project.WebPath)
	if err != nil {
		logrus.Errorf("%s 文件不存在", site.Project.WebPath)
		return errors.New(fmt.Sprintf("%s 文件不存在", site.Project.WebPath))
	}

	doc, err := goquery.NewDocumentFromReader(file)
	if err != nil {
		logrus.Errorf("%s 文件解析失败", site.Project.WebPath)
		return errors.New("文件解析失败")
	}

	// 开始修改
	if site.Project.Title != "" {
		doc.Find("title").SetText(site.Project.Title)
	}

	if site.Project.Icon != "" {
		selection := doc.Find("link[rel=\"icon\"]")
		if selection.Length() > 0 {
			// 有就修改
			doc.Find("link[rel=\"icon\"]").SetAttr("href", site.Project.Icon)
		} else {
			// 没有就创建
			doc.Find("head").AppendHtml(fmt.Sprintf("<link rel=\"icon\" href=\"%s\">\n", site.Project.Icon))
		}
	}

	if site.Seo.Keywords != "" {
		selection := doc.Find("meta[name=\"keywords\"]")
		if selection.Length() > 0 {
			// 有就修改
			doc.Find("meta[name=\"keywords\"]").SetAttr("content", site.Seo.Keywords)
		} else {
			// 没有就创建
			doc.Find("head").AppendHtml(fmt.Sprintf("<meta name=\"keywords\" content=\"%s\">\n", site.Seo.Keywords))
		}
	}

	if site.Seo.Description != "" {
		selection := doc.Find("meta[name=\"description\"]")
		if selection.Length() > 0 {
			// 有就修改
			doc.Find("meta[name=\"description\"]").SetAttr("content", site.Seo.Description)
		} else {
			// 没有就创建
			doc.Find("head").AppendHtml(fmt.Sprintf("<meta name=\"description\" content=\"%s\">\n", site.Seo.Description))
		}
	}

	// 生成 html
	html, err := doc.Html()
	if err != nil {
		logrus.Errorf("生成 html 错误: %s", err)
		return errors.New("生成 html 错误")
	}

	// 写入前端文件
	err = os.WriteFile(site.Project.WebPath, []byte(html), 0666)
	if err != nil {
		logrus.Errorf("html 文件写入失败: %s", err)
		return errors.New("html 文件写入失败")
	}
	return nil
}
