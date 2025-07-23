package crawlerservice

import "fmt"

// ArxivCategory 论文类别枚举
type ArxivCategory int

const (
	CategoryAI ArxivCategory = iota + 1
	CategoryAstroPhysics
	CategoryHighEnergyPhysics
	CategoryQuantumPhysics
	CategoryMathematics
	CategoryComputerScience
	CategoryPhysics
)

// ArxivCategoryConfig 论文类别配置
type ArxivCategoryConfig struct {
	URL         string `json:"url"`         // ArXiv URL
	ChineseName string `json:"chineseName"` // 中文名称
	EnglishName string `json:"englishName"` // 英文名称
	Code        string `json:"code"`        // 类别代码
}

// categoryConfigs 各类别的配置映射
var categoryConfigs = map[ArxivCategory]ArxivCategoryConfig{
	CategoryAI: {
		URL:         "https://arxiv.org/list/cs.AI/new",
		ChineseName: "人工智能",
		EnglishName: "Artificial Intelligence",
		Code:        "cs.AI",
	},
	CategoryAstroPhysics: {
		URL:         "https://arxiv.org/list/astro-ph/new",
		ChineseName: "天体物理学",
		EnglishName: "Astrophysics",
		Code:        "astro-ph",
	},
	CategoryHighEnergyPhysics: {
		URL:         "https://arxiv.org/list/hep-ex/new",
		ChineseName: "高能物理实验",
		EnglishName: "High Energy Physics - Experiment",
		Code:        "hep-ex",
	},
	CategoryQuantumPhysics: {
		URL:         "https://arxiv.org/list/quant-ph/new",
		ChineseName: "量子物理",
		EnglishName: "Quantum Physics",
		Code:        "quant-ph",
	},
	CategoryMathematics: {
		URL:         "https://arxiv.org/list/math/new",
		ChineseName: "数学",
		EnglishName: "Mathematics",
		Code:        "math",
	},
	CategoryComputerScience: {
		URL:         "https://arxiv.org/list/cs/new",
		ChineseName: "计算机科学",
		EnglishName: "Computer Science",
		Code:        "cs",
	},
	CategoryPhysics: {
		URL:         "https://arxiv.org/list/physics/new",
		ChineseName: "物理学",
		EnglishName: "Physics",
		Code:        "physics",
	},
}

// String 返回类别的中文名称
func (c ArxivCategory) String() string {
	if config, exists := categoryConfigs[c]; exists {
		return config.ChineseName
	}
	return "未知类别"
}

// GetURL 获取类别对应的ArXiv URL
func (c ArxivCategory) GetURL() string {
	if config, exists := categoryConfigs[c]; exists {
		return config.URL
	}
	return ""
}

// GetEnglishName 获取类别的英文名称
func (c ArxivCategory) GetEnglishName() string {
	if config, exists := categoryConfigs[c]; exists {
		return config.EnglishName
	}
	return "Unknown Category"
}

// GetCode 获取类别代码
func (c ArxivCategory) GetCode() string {
	if config, exists := categoryConfigs[c]; exists {
		return config.Code
	}
	return ""
}

// GetConfig 获取完整的类别配置
func (c ArxivCategory) GetConfig() ArxivCategoryConfig {
	if config, exists := categoryConfigs[c]; exists {
		return config
	}
	return ArxivCategoryConfig{}
}

// GetAllCategories 获取所有可用的类别
func GetAllCategories() []ArxivCategory {
	categories := make([]ArxivCategory, 0, len(categoryConfigs))
	for category := range categoryConfigs {
		categories = append(categories, category)
	}
	return categories
}

// GetCategoryByCode 根据代码获取类别
func GetCategoryByCode(code string) (ArxivCategory, error) {
	for category, config := range categoryConfigs {
		if config.Code == code {
			return category, nil
		}
	}
	return 0, fmt.Errorf("未找到代码为 %s 的类别", code)
}

var (
	// ArxivCrawlerService 全局ArXiv爬虫服务实例
	ArxivCrawlerService = NewArxivCrawler()
)
