package autogen_service

import (
	"blogX_server/service/article_auto_generate/crawler_service"
)

// AutogenService AI自动分析服务
type AutogenService struct {
	crawlerService *crawler_service.ArxivCrawler
}

// NewAutogenService 创建新的自动分析服务实例
func NewAutogenService() *AutogenService {
	return &AutogenService{
		crawlerService: crawler_service.NewArxivCrawler(),
	}
}

var (
	// AutogenServiceInstance 全局自动分析服务实例
	AutogenServiceInstance = NewAutogenService()
)
