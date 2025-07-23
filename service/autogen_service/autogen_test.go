package autogen_service

import (
	"testing"

	"blogX_server/core"
	"blogX_server/flags"
	"blogX_server/global"
	"blogX_server/service/crawler_service"

	"github.com/sirupsen/logrus"
)

// setupForTest 初始化测试环境
func setupForTest() {
	flags.Parse()
	flags.FlagOptions.File = "../../settings.yaml"
	global.Config = core.ReadConf()
	core.InitLogrus()
	global.Redis = core.InitRedis()
}

// TestNewArchitecture 测试新架构：爬虫+AI分析缓存
func TestNewArchitecture(t *testing.T) {
	// 初始化
	setupForTest()

	service := NewAutogenService()

	logrus.Info("=== 测试新架构：爬虫数据实时获取，AI结果缓存 ===")

	// 测试实时爬取+AI分析（带缓存）
	topPapers, err := service.AnalyzePapersForWriting(crawler_service.CategoryAI, 10, 3)
	if err != nil {
		t.Errorf("分析失败: %v", err)
		return
	}

	if len(topPapers) == 0 {
		t.Error("没有获取到分析结果")
		return
	}

	logrus.Infof("✅ 第一次分析完成，获得 %d 篇高分论文", len(topPapers))

	// 显示分析结果
	for i, paper := range topPapers {
		logrus.Infof("【论文 %d】%s (评分: %d)", i+1, paper.Title, paper.Score)
		logrus.Infof("PDF: %s", paper.PdfURL)
		logrus.Infof("HTML: %s", paper.HtmlURL)
		logrus.Infof("标签: %v", paper.Tags)
		logrus.Info("---")
	}

	// 测试缓存功能 - 再次调用应该使用缓存
	logrus.Info("=== 测试AI分析缓存功能 ===")

	// 获取缓存统计
	stats, err := service.GetCacheStats()
	if err == nil {
		logrus.Infof("缓存统计: %+v", stats)
	}

	// 再次分析相同论文（应该大部分使用缓存）
	topPapers2, err := service.AnalyzePapersForWriting(crawler_service.CategoryAI, 10, 3)
	if err != nil {
		t.Errorf("第二次分析失败: %v", err)
		return
	}

	logrus.Infof("✅ 第二次分析完成，获得 %d 篇高分论文（应该有缓存命中）", len(topPapers2))

	// 测试缓存清理
	logrus.Info("=== 测试缓存清理 ===")
	err = service.ClearAnalysisCache()
	if err != nil {
		logrus.Errorf("清理缓存失败: %v", err)
	} else {
		logrus.Info("✅ 缓存清理成功")
	}
}

// TestMultiCategory 测试多类别论文分析
func TestMultiCategory(t *testing.T) {
	setupForTest()

	service := NewAutogenService()

	logrus.Info("=== 测试多类别论文分析 ===")

	categories := []crawler_service.ArxivCategory{
		crawler_service.CategoryAI,
		crawler_service.CategoryQuantumPhysics,
	}

	for _, category := range categories {
		logrus.Infof("开始分析 %s 类别论文...", category.String())

		topPapers, err := service.AnalyzePapersForWriting(category, 5, 2)
		if err != nil {
			t.Logf("分析 %s 失败: %v", category.String(), err)
			continue
		}

		logrus.Infof("✅ %s 类别分析完成，获得 %d 篇高分论文", category.String(), len(topPapers))

		for i, paper := range topPapers {
			logrus.Infof("[%s %d] %s (评分: %d)", category.String(), i+1, paper.Title, paper.Score)
		}
	}
}

// TestCacheManagement 专门测试缓存管理功能
func TestCacheManagement(t *testing.T) {
	setupForTest()

	service := NewAutogenService()

	logrus.Info("=== 测试缓存管理功能 ===")

	// 1. 清理旧缓存
	err := service.ClearAnalysisCache()
	if err != nil {
		t.Errorf("清理缓存失败: %v", err)
		return
	}
	logrus.Info("✅ 清理旧缓存成功")

	// 2. 获取初始统计
	stats, err := service.GetCacheStats()
	if err != nil {
		t.Errorf("获取缓存统计失败: %v", err)
		return
	}
	logrus.Infof("初始缓存统计: %+v", stats)

	// 3. 执行分析，产生新缓存
	_, err = service.AnalyzePapersForWriting(crawler_service.CategoryAI, 3, 2)
	if err != nil {
		t.Errorf("分析失败: %v", err)
		return
	}
	logrus.Info("✅ 分析完成，应该产生了新缓存")

	// 4. 获取分析后统计
	stats2, err := service.GetCacheStats()
	if err != nil {
		t.Errorf("获取缓存统计失败: %v", err)
		return
	}
	logrus.Infof("分析后缓存统计: %+v", stats2)

	// 5. 验证缓存增加
	if stats2["total_cached"].(int) > stats["total_cached"].(int) {
		logrus.Info("✅ 缓存正常增加")
	} else {
		t.Error("❌ 缓存没有增加")
	}
}
