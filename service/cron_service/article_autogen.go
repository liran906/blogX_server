// Path: ./service/cron_service/article_autogen.go

package cron_service

import (
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/ctype"
	"blogX_server/models/enum"
	"blogX_server/service/autogen_service"
	"blogX_server/service/crawler_service"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

func AutoGenerateArxivAbstract() {
	limit := global.Config.Site.AutoGen.Limit
	top := global.Config.Site.AutoGen.Top
	for _, code := range global.Config.Site.AutoGen.Categories {
		c, err := crawler_service.GetCategoryByCode(code)
		if err != nil {
			logrus.Errorf("get arxiv category error: %v", err)
			continue
		}
		autoGenerateArxivAbstract(c, limit, top)
	}
}

func autoGenerateArxivAbstract(category crawler_service.ArxivCategory, limit, topN int) {
	// 1. 指定类别爬取
	// 指定特定类别
	crawler := crawler_service.NewArxivCrawlerWithCategory(category)
	papers, err := crawler.CrawlRecentPapers()
	if err != nil {
		logrus.Errorf("crawl papers error: %v", err)
		return
	}
	if limit > 0 {
		papers = papers[:limit]
	}

	// 2. 分析筛选
	service := autogen_service.NewAutogenService()
	topPapers, _ := service.AnalyzePapersFromList(papers, topN)
	content := autogen_service.FormatAnalysisReport(topPapers, category.String())

	err = articleGen(content, category.String())
	if err != nil {
		logrus.Errorf("article auto gen error: %v", err)
		return
	}
	//service.ClearAnalysisCache()
}

func articleGen(content, category string) error {
	uid := global.Config.Site.AutoGen.UserID

	// 取分类
	var cat models.CategoryModel
	if category != "" {
		err := global.DB.Take(&cat, "name = ? and user_id = ?", category, uid).Error
		if err != nil {
			if err.Error() == gorm.ErrRecordNotFound.Error() {
				cat.Name = category
				cat.UserID = uid
				err = global.DB.Create(&cat).Error
				if err != nil {
					logrus.Errorf("创建新分类失败 %v", err)
				}
			} else {
				logrus.Errorf("文章分类错误 %v", err)
			}
		}
	}

	nowDate := time.Now().Format("01月02日")
	abstract := fmt.Sprintf("每天让 AI 助手，从特定领域 百篇最新论文 中整理挑选最好的 30 篇，供您阅读")

	var article = models.ArticleModel{
		Title:          nowDate + " " + category,
		Abstract:       abstract,
		CoverURL:       "/uploads/images/9742aaccce6aaf3078e1f9df8bcc222d.png",
		Content:        content,
		CategoryID:     &cat.ID,
		Tags:           ctype.List{category},
		OpenForComment: true,
		UserID:         uid,
		Status:         enum.ArticleStatusPublish, // 自动免审
	}

	// 入库
	err := global.DB.Create(&article).Error
	if err != nil {
		logrus.Errorf("文章自动发布失败，%v", err)
		return err
	}
	logrus.Info("文章自动生成发布成功")
	return nil
}
