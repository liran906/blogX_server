package crawler_service

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"blogX_server/core"
	"blogX_server/flags"
	"blogX_server/global"

	"github.com/sirupsen/logrus"
)

// setupRedisForTest 初始化Redis连接用于测试
func setupRedisForTest() {
	flags.Parse() // 解析命令行

	// 设置配置文件路径为项目根目录
	flags.FlagOptions.File = "../../settings.yaml"

	global.Config = core.ReadConf() // 读取配置文件
	core.InitLogrus()               // 初始化日志文件
	global.Redis = core.InitRedis() // 连接 redis
}

// cleanupRedisForTest 清理测试数据
func cleanupRedisForTest() {
	if global.Redis != nil {
		// 删除测试相关的Redis键
		keys := []string{
			RedisListKey,         // arxiv:papers:list
			RedisKeyPrefix + "*", // arxiv:papers:*
		}

		for _, key := range keys {
			if strings.Contains(key, "*") {
				// 处理通配符键
				result, err := global.Redis.Keys(key).Result()
				if err == nil {
					for _, k := range result {
						global.Redis.Del(k)
					}
				}
			} else {
				global.Redis.Del(key)
			}
		}
		logrus.Info("已清理测试Redis数据")
	}
}

// TestArxivCrawler 测试ArXiv爬虫功能
func TestArxivCrawler(t *testing.T) {
	// 初始化Redis
	setupRedisForTest()
	// defer cleanupRedisForTest() // 确保测试完成后清理数据

	logrus.Info("=== 第一步：调试HTML结构 ===")
	err := DebugArxivHTML()
	if err != nil {
		t.Errorf("调试HTML失败: %v", err)
		return
	}

	logrus.Info("=== 第二步：尝试爬取论文 ===")
	crawler := NewArxivCrawler()

	// 1. 测试爬取论文列表
	papers, err := crawler.CrawlRecentAIPapers()
	if err != nil {
		t.Errorf("爬取失败: %v", err)
		return
	}

	logrus.Infof("爬取到 %d 篇论文", len(papers))

	if len(papers) == 0 {
		t.Log("警告：没有爬取到论文，可能需要调整HTML解析逻辑")
		t.Log("请查看上面的调试信息来了解实际的HTML结构")
		return
	}

	// 输出前3篇论文信息
	for i, paper := range papers {
		if i >= 3 {
			break
		}
		fmt.Printf("\n=== 论文 %d ===\n", i+1)
		fmt.Printf("ArXiv ID: %s\n", paper.ArxivID)
		fmt.Printf("标题: %s\n", paper.Title)
		fmt.Printf("作者: %s\n", paper.Authors)
		fmt.Printf("PDF链接: %s\n", paper.PdfURL)
		fmt.Printf("HTML链接: %s\n", paper.HtmlURL)
		fmt.Printf("爬取时间: %s\n", paper.CrawlTime)
	}

	// 2. 测试爬取单篇论文摘要
	if len(papers) > 0 {
		firstPaper := papers[0]
		logrus.Infof("尝试爬取论文摘要: %s", firstPaper.ArxivID)

		detailedPaper, err := crawler.CrawlPaperDetails(firstPaper.ArxivID)
		if err != nil {
			logrus.Errorf("爬取摘要失败: %v", err)
		} else {
			fmt.Printf("\n=== 详细论文信息 ===\n")
			fmt.Printf("标题: %s\n", detailedPaper.Title)
			fmt.Printf("作者: %s\n", detailedPaper.Authors)
			fmt.Printf("摘要: %s\n", detailedPaper.Abstract)
		}
	}

	// 3. 测试Redis存储和读取
	logrus.Info("=== 第三步：测试Redis功能 ===")

	// 测试保存到Redis并设置过期时间
	err = crawler.SaveToRedis(papers)
	if err != nil {
		logrus.Errorf("保存到Redis失败: %v", err)
	} else {
		logrus.Info("成功保存到Redis")

		// 为Redis键设置过期时间（1小时）
		if global.Redis != nil {
			global.Redis.Expire(RedisListKey, time.Hour)
			logrus.Info("已为Redis键设置1小时过期时间")
		}
	}

	// 测试从Redis读取
	storedPapers, err := crawler.GetFromRedis(10)
	if err != nil {
		logrus.Errorf("从Redis读取失败: %v", err)
	} else {
		logrus.Infof("从Redis成功读取到 %d 篇论文", len(storedPapers))
		if len(storedPapers) > 0 {
			fmt.Printf("Redis存储的第一篇论文: %s - %s\n",
				storedPapers[0].ArxivID, storedPapers[0].Title)
		}
	}

	// 4. 测试搜索功能
	if len(papers) > 0 {
		logrus.Info("=== 第四步：测试搜索功能 ===")
		searchResults, err := SearchPapers("learning", 5)
		if err != nil {
			logrus.Errorf("搜索失败: %v", err)
		} else {
			logrus.Infof("搜索 'learning' 找到 %d 篇相关论文", len(searchResults))
		}
	}
}

// TestDebugOnly 仅调试HTML结构
func TestDebugOnly(t *testing.T) {
	setupRedisForTest()
	// defer cleanupRedisForTest()

	err := DebugArxivHTML()
	if err != nil {
		t.Errorf("调试失败: %v", err)
	}
}

// TestRedisOperations 专门测试Redis操作
func TestRedisOperations(t *testing.T) {
	setupRedisForTest()
	// defer cleanupRedisForTest()

	if global.Redis == nil {
		t.Skip("Redis未连接，跳过Redis测试")
	}

	crawler := NewArxivCrawler()

	// 创建测试数据
	testPapers := []ArxivPaper{
		{
			ArxivID:      "arXiv:test.001",
			Title:        "Test Paper 1",
			Authors:      "Test Author",
			Abstract:     "This is a test abstract for testing Redis functionality.",
			PdfURL:       "https://arxiv.org/pdf/test.001.pdf",
			HtmlURL:      "https://arxiv.org/html/test.001",
			Category:     CategoryAI,
			CategoryName: "人工智能",
			CrawlTime:    time.Now().Format("2006-01-02 15:04:05"),
		},
		{
			ArxivID:      "arXiv:test.002",
			Title:        "Test Paper 2",
			Authors:      "Another Author",
			Abstract:     "Another test abstract with machine learning keywords.",
			PdfURL:       "https://arxiv.org/pdf/test.002.pdf",
			HtmlURL:      "https://arxiv.org/html/test.002",
			Category:     CategoryAI,
			CategoryName: "人工智能",
			CrawlTime:    time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	// 测试保存
	err := crawler.SaveToRedis(testPapers)
	if err != nil {
		t.Errorf("保存到Redis失败: %v", err)
		return
	}
	logrus.Info("✅ Redis保存测试通过")

	// 设置过期时间
	global.Redis.Expire(RedisListKey, 10*time.Minute)

	// 测试读取
	retrievedPapers, err := crawler.GetFromRedis(10)
	if err != nil {
		t.Errorf("从Redis读取失败: %v", err)
		return
	}

	if len(retrievedPapers) != len(testPapers) {
		t.Errorf("期望读取 %d 篇论文，实际读取 %d 篇", len(testPapers), len(retrievedPapers))
		return
	}
	logrus.Info("✅ Redis读取测试通过")

	// 测试搜索
	searchResults, err := SearchPapers("machine learning", 5)
	if err != nil {
		t.Errorf("搜索失败: %v", err)
		return
	}

	if len(searchResults) == 0 {
		t.Log("警告：搜索没有找到匹配的论文")
	} else {
		logrus.Infof("✅ 搜索测试通过，找到 %d 篇相关论文", len(searchResults))
	}
}

// GetPapersAsJSON 获取论文列表的JSON格式
func GetPapersAsJSON(limit int) (string, error) {
	crawler := NewArxivCrawler()

	// 先尝试从Redis获取
	papers, err := crawler.GetFromRedis(limit)
	if err != nil || len(papers) == 0 {
		// 如果Redis没有数据，重新爬取
		logrus.Info("Redis中没有数据，开始爬取...")
		papers, err = crawler.CrawlAndSave()
		if err != nil {
			return "", err
		}

		// 限制返回数量
		if len(papers) > limit {
			papers = papers[:limit]
		}
	}

	// 转换为JSON
	jsonData, err := json.MarshalIndent(papers, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化JSON失败: %v", err)
	}

	return string(jsonData), nil
}

// GetRecommendationData 获取用于推荐的论文数据
func GetRecommendationData() ([]ArxivPaper, error) {
	crawler := NewArxivCrawler()

	// 先尝试从Redis获取
	papers, err := crawler.GetFromRedis(50) // 获取50篇用于推荐
	if err != nil || len(papers) == 0 {
		// 如果Redis没有数据，重新爬取
		logrus.Info("Redis中没有数据，开始爬取...")
		papers, err = crawler.CrawlAndSave()
		if err != nil {
			return nil, err
		}
	}

	// 过滤出有摘要的论文用于推荐
	var validPapers []ArxivPaper
	for _, paper := range papers {
		if paper.Abstract != "" && len(paper.Abstract) > 50 { // 确保摘要有足够内容
			validPapers = append(validPapers, paper)
		}
	}

	logrus.Infof("获取到 %d 篇有效论文用于推荐", len(validPapers))
	return validPapers, nil
}

// SearchPapers 在已爬取的论文中搜索
func SearchPapers(keyword string, limit int) ([]ArxivPaper, error) {
	crawler := NewArxivCrawler()

	// 从Redis获取所有论文
	papers, err := crawler.GetFromRedis(1000) // 获取足够多的论文
	if err != nil || len(papers) == 0 {
		return nil, fmt.Errorf("没有可搜索的论文数据")
	}

	var results []ArxivPaper
	keyword = strings.ToLower(keyword) // 转为小写

	for _, paper := range papers {
		// 在标题、作者、摘要中搜索关键词
		if containsIgnoreCase(paper.Title, keyword) ||
			containsIgnoreCase(paper.Authors, keyword) ||
			containsIgnoreCase(paper.Abstract, keyword) {
			results = append(results, paper)

			if len(results) >= limit {
				break
			}
		}
	}

	logrus.Infof("搜索关键词 '%s' 找到 %d 篇论文", keyword, len(results))
	return results, nil
}

// containsIgnoreCase 忽略大小写的字符串包含检查
func containsIgnoreCase(text, keyword string) bool {
	if text == "" || keyword == "" {
		return false
	}

	// 忽略大小写的字符串包含检查
	return strings.Contains(strings.ToLower(text), strings.ToLower(keyword))
}
