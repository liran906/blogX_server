package crawlerservice

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
)

// TestCategoryString 测试类别的String()方法
func TestCategoryString(t *testing.T) {
	testCases := []struct {
		category ArxivCategory
		expected string
	}{
		{CategoryAI, "人工智能"},
		{CategoryAstroPhysics, "天体物理学"},
		{CategoryHighEnergyPhysics, "高能物理实验"},
		{CategoryQuantumPhysics, "量子物理"},
		{CategoryMathematics, "数学"},
		{CategoryComputerScience, "计算机科学"},
		{CategoryPhysics, "物理学"},
	}

	for _, tc := range testCases {
		result := tc.category.String()
		if result != tc.expected {
			t.Errorf("类别 %v: 期望 %s, 得到 %s", tc.category, tc.expected, result)
		}
		fmt.Printf("✅ %s -> %s\n", tc.category.GetCode(), result)
	}
}

// TestCategoryMethods 测试类别的各种方法
func TestCategoryMethods(t *testing.T) {
	category := CategoryAstroPhysics

	fmt.Printf("=== 测试类别方法 ===\n")
	fmt.Printf("类别代码: %s\n", category.GetCode())
	fmt.Printf("中文名称: %s\n", category.String())
	fmt.Printf("英文名称: %s\n", category.GetEnglishName())
	fmt.Printf("爬取URL: %s\n", category.GetURL())

	config := category.GetConfig()
	fmt.Printf("完整配置: %+v\n", config)
}

// TestGetCategoryByCode 测试根据代码获取类别
func TestGetCategoryByCode(t *testing.T) {
	testCases := map[string]ArxivCategory{
		"cs.AI":    CategoryAI,
		"astro-ph": CategoryAstroPhysics,
		"hep-ex":   CategoryHighEnergyPhysics,
		"quant-ph": CategoryQuantumPhysics,
		"math":     CategoryMathematics,
		"cs":       CategoryComputerScience,
		"physics":  CategoryPhysics,
	}

	for code, expectedCategory := range testCases {
		category, err := GetCategoryByCode(code)
		if err != nil {
			t.Errorf("根据代码 %s 获取类别失败: %v", code, err)
			continue
		}

		if category != expectedCategory {
			t.Errorf("代码 %s: 期望类别 %v, 得到 %v", code, expectedCategory, category)
			continue
		}

		fmt.Printf("✅ 代码 %s -> %s\n", code, category.String())
	}

	// 测试无效代码
	_, err := GetCategoryByCode("invalid")
	if err == nil {
		t.Error("期望无效代码返回错误，但没有返回")
	}
	fmt.Printf("✅ 无效代码正确返回错误: %v\n", err)
}

// TestMultiCategoryCrawl 测试多类别爬虫功能
func TestMultiCategoryCrawl(t *testing.T) {
	logrus.Info("=== 测试多类别爬虫功能 ===")

	// 测试各个类别的爬虫
	categories := []ArxivCategory{
		CategoryAI,
		CategoryAstroPhysics,
		// CategoryHighEnergyPhysics, // 注释掉一些类别以减少测试时间
	}

	for _, category := range categories {
		logrus.Infof("测试爬取 %s 论文...", category.String())

		papers, err := CrawlPapersByCategory(category)
		if err != nil {
			t.Logf("爬取 %s 论文失败: %v", category.String(), err)
			continue
		}

		logrus.Infof("✅ 成功爬取 %s 论文 %d 篇", category.String(), len(papers))

		// 显示前2篇论文
		for i, paper := range papers {
			if i >= 2 {
				break
			}
			fmt.Printf("\n--- %s 论文 %d ---\n", category.String(), i+1)
			fmt.Printf("ArXiv ID: %s\n", paper.ArxivID)
			fmt.Printf("标题: %s\n", paper.Title)
			fmt.Printf("作者: %s\n", paper.Authors)
			fmt.Printf("类别: %s (%s)\n", paper.CategoryName, paper.Category.GetCode())
			fmt.Printf("PDF链接: %s\n", paper.PdfURL)
		}
	}
}

// TestConvenienceMethods 测试便利方法
func TestConvenienceMethods(t *testing.T) {
	logrus.Info("=== 测试便利方法 ===")

	// 只测试一个类别以节省时间
	logrus.Info("测试天体物理学便利方法...")
	papers, err := CrawlAstrophysicsPapers()
	if err != nil {
		t.Logf("爬取天体物理学论文失败: %v", err)
		return
	}

	logrus.Infof("✅ 便利方法成功爬取天体物理学论文 %d 篇", len(papers))

	// 验证类别信息
	if len(papers) > 0 {
		paper := papers[0]
		if paper.Category != CategoryAstroPhysics {
			t.Errorf("期望类别 %v, 得到 %v", CategoryAstroPhysics, paper.Category)
		}
		if paper.CategoryName != "天体物理学" {
			t.Errorf("期望类别名称 '天体物理学', 得到 '%s'", paper.CategoryName)
		}
		fmt.Printf("✅ 类别信息正确: %s (%s)\n", paper.CategoryName, paper.Category.GetCode())
	}
}

// TestAllCategories 测试获取所有类别
func TestAllCategories(t *testing.T) {
	categories := GetAllCategories()

	fmt.Printf("=== 所有可用类别 ===\n")
	for _, category := range categories {
		fmt.Printf("- %s (%s): %s\n",
			category.String(),
			category.GetCode(),
			category.GetEnglishName())
	}

	if len(categories) != 7 {
		t.Errorf("期望7个类别，得到 %d 个", len(categories))
	}
}
