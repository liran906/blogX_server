package autogen_service

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"blogX_server/global"
	"blogX_server/service/ai_service"
	"blogX_server/service/crawler_service"

	"github.com/sirupsen/logrus"
)

const (
	// AI分析结果Redis缓存键前缀
	AnalysisResultPrefix = "ai_analysis:"
	// 分析结果缓存时间（7天）
	AnalysisCacheExpiration = 7 * 24 * time.Hour
)

// PaperAnalysisResult AI分析结果结构体
type PaperAnalysisResult struct {
	ArxivID          string   `json:"arxivId"`          // 原始ArXiv ID
	Title            string   `json:"title"`            // 论文标题
	Authors          string   `json:"authors"`          // 作者
	PublishedDate    string   `json:"publishedDate"`    // 发表时间
	Abstract         string   `json:"abstract"`         // AI生成的中文摘要
	Score            int      `json:"score"`            // 科研价值评分(0-100)
	Justification    string   `json:"just"`             // 评分理由
	Tags             []string `json:"tags"`             // 主题标签
	AnalyzedAt       string   `json:"analyzedAt"`       // 分析时间
	OriginalAbstract string   `json:"originalAbstract"` // 原始英文摘要
	PdfURL           string   `json:"pdfUrl"`           // PDF链接
	HtmlURL          string   `json:"htmlUrl"`          // HTML链接
}

// AIAnalysisResponse AI返回的分析结果
type AIAnalysisResponse struct {
	Abstract string   `json:"abstract"` // 中文摘要
	Score    int      `json:"score"`    // 评分
	Just     string   `json:"just"`     // 评分理由
	Tags     []string `json:"tags"`     // 标签
}

// AnalyzePaper 分析单篇论文（自动处理缓存）
func (s *AutogenService) AnalyzePaper(paper *crawler_service.ArxivPaper) (*PaperAnalysisResult, error) {
	if paper == nil {
		return nil, fmt.Errorf("论文数据为空")
	}

	// 1. 首先检查缓存
	if cached := s.getAnalysisFromCache(paper.ArxivID); cached != nil {
		logrus.Debugf("使用缓存的分析结果: %s", paper.ArxivID)
		return cached, nil
	}

	logrus.Infof("开始AI分析论文: %s", paper.ArxivID)

	// 2. 构建AI分析的输入文本
	inputText := fmt.Sprintf("标题：%s\n内容：%s", paper.Title, paper.Abstract)

	// 3. 调用AI服务进行分析
	aiResponse, err := ai_service.Autogen(inputText)
	if err != nil {
		logrus.Errorf("AI分析失败: %v", err)
		return nil, fmt.Errorf("AI分析失败: %v", err)
	}

	// 4. 清理AI返回的结果，去除可能的代码块标记
	cleanResponse := strings.TrimSpace(aiResponse)
	cleanResponse = strings.TrimPrefix(cleanResponse, "```json")
	cleanResponse = strings.TrimPrefix(cleanResponse, "```")
	cleanResponse = strings.TrimSuffix(cleanResponse, "```")
	cleanResponse = strings.TrimSpace(cleanResponse)

	// 5. 解析AI返回的JSON结果
	var aiResult AIAnalysisResponse
	err = json.Unmarshal([]byte(cleanResponse), &aiResult)
	if err != nil {
		logrus.Errorf("AI返回结果解析失败: %v, 原始结果: %s", err, aiResponse)
		return nil, fmt.Errorf("AI返回结果解析失败: %v", err)
	}

	// 6. 构建最终结果
	result := &PaperAnalysisResult{
		ArxivID:          paper.ArxivID,
		Title:            paper.Title,
		Authors:          paper.Authors,
		PublishedDate:    paper.PublishedDate,
		Abstract:         aiResult.Abstract,
		Score:            aiResult.Score,
		Justification:    aiResult.Just,
		Tags:             aiResult.Tags,
		AnalyzedAt:       time.Now().Format("2006-01-02 15:04:05"),
		OriginalAbstract: paper.Abstract,
		PdfURL:           fmt.Sprintf("https://arxiv.org/pdf/%s.pdf", paper.ArxivID),
		HtmlURL:          fmt.Sprintf("https://arxiv.org/abs/%s", paper.ArxivID),
	}

	// 7. 自动保存到缓存
	s.saveAnalysisToCache(result)

	logrus.Infof("论文 %s 分析完成，评分: %d（已缓存）", paper.ArxivID, result.Score)
	return result, nil
}

// AnalyzePapers 批量分析论文（自动缓存）
func (s *AutogenService) AnalyzePapers(papers []crawler_service.ArxivPaper) ([]*PaperAnalysisResult, error) {
	if len(papers) == 0 {
		return nil, fmt.Errorf("论文列表为空")
	}

	results := make([]*PaperAnalysisResult, 0, len(papers))

	logrus.Infof("开始批量分析 %d 篇论文", len(papers))

	for i, paper := range papers {
		result, err := s.AnalyzePaper(&paper)
		if err != nil {
			logrus.Errorf("分析第 %d 篇论文失败: %v", i+1, err)
			// 继续处理其他论文，而不是中断整个过程
			continue
		}

		results = append(results, result)

		// 为了避免API限制，在每次请求之间添加短暂延迟
		if i < len(papers)-1 {
			time.Sleep(1 * time.Second)
		}
	}

	logrus.Infof("批量分析完成，成功分析 %d/%d 篇论文", len(results), len(papers))
	return results, nil
}

// AnalyzePapersWithCache 带缓存的批量论文分析（简化版）
func (s *AutogenService) AnalyzePapersWithCache(papers []crawler_service.ArxivPaper) ([]*PaperAnalysisResult, error) {
	logrus.Infof("开始分析 %d 篇论文（自动缓存）", len(papers))

	var results []*PaperAnalysisResult

	// 直接调用AnalyzePaper，它会自动处理缓存
	for _, paper := range papers {
		result, err := s.AnalyzePaper(&paper)
		if err != nil {
			logrus.Errorf("分析论文失败 %s: %v", paper.ArxivID, err)
			continue // 跳过失败的论文，继续分析其他论文
		}
		results = append(results, result)
	}

	logrus.Infof("分析完成，总共 %d 篇论文", len(results))
	return results, nil
}

// GetTopScoredPapers 获取评分最高的论文
func GetTopScoredPapers(results []*PaperAnalysisResult, topN int) []*PaperAnalysisResult {
	// 简单的排序，按评分降序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if topN <= 0 {
		return results
	}

	// 确保topN不超过数组长度
	if topN > len(results) {
		topN = len(results)
	}

	return results[:topN]
}

// FormatAnalysisReport 格式化分析报告为 Markdown 格式
func FormatAnalysisReport(results []*PaperAnalysisResult, category string) string {
	if len(results) == 0 {
		return "## 📊 AI论文分析报告\n\n> ❌ 没有分析结果"
	}

	var body strings.Builder

	// 报告头部
	body.WriteString("### 每天让 AI 助手，从特定领域 百篇最新论文 中整理挑选最好的 30 篇，供您阅读\n\n")
	body.WriteString("### 今日概览\n\n")
	body.WriteString(fmt.Sprintf("**学术领域**: %s  \n", category))
	body.WriteString(fmt.Sprintf("**生成时间**: %s  \n", time.Now().Format("2006-01-02 15:04:05")))
	body.WriteString(fmt.Sprintf("**分析数量**: %d 篇论文  \n", len(results)))

	// 评分概览
	var scores []int
	for _, result := range results {
		scores = append(scores, result.Score)
	}
	avgScore := calculateAverage(scores)
	maxScore := findMax(scores)
	minScore := findMin(scores)

	body.WriteString(fmt.Sprintf("- **最高分**: `%d`  \n", maxScore))
	body.WriteString(fmt.Sprintf("- **平均分**: `%.1f`  \n", avgScore))
	body.WriteString(fmt.Sprintf("- **最低分**: `%d`  \n\n", minScore))

	body.WriteString("---\n\n")

	// 📝 论文详情
	for i, paper := range results {
		// 👥 作者名称截断处理
		authors := paper.Authors
		if len(authors) > 30 {
			authors = authors[:30] + "..."
		}

		body.WriteString(fmt.Sprintf("### %s\n\n", paper.Title))
		body.WriteString(fmt.Sprintf("**作者**: %s  \n", authors))
		// 📋 标签展示
		if len(paper.Tags) > 0 {
			body.WriteString("**标签**: ")
			for j, tag := range paper.Tags {
				body.WriteString(fmt.Sprintf("`%s`", tag))
				if j < len(paper.Tags)-1 {
					body.WriteString(" ")
				}
			}
			body.WriteString("  \n")
		}
		body.WriteString(fmt.Sprintf("**分析时间**: %s  \n", paper.AnalyzedAt))
		body.WriteString(fmt.Sprintf("**论文源**: [ ArXiv ](%s) | [ PDF ](%s)  \n", paper.HtmlURL, paper.PdfURL))
		body.WriteString(fmt.Sprintf("**本站评分**: `%d/100`\n", paper.Score))
		body.WriteString(fmt.Sprintf("**本站分析**: %s  \n", paper.Justification))

		body.WriteString(fmt.Sprintf("**AI摘要**: %s\n\n", paper.Abstract))

		// 分隔符（最后一篇不添加）
		if i < len(results)-1 {
			body.WriteString("---\n\n")
		}
	}

	return body.String()
}

// AnalyzePapersForWriting 专门用于写文章的论文分析（推荐使用）
func (s *AutogenService) AnalyzePapersForWriting(category crawler_service.ArxivCategory, limit int, topN int) ([]*PaperAnalysisResult, error) {
	// 1. 实时爬取最新论文
	crawler := crawler_service.NewArxivCrawlerWithCategory(category)
	papers, err := crawler.CrawlRecentPapers()
	if err != nil {
		return nil, fmt.Errorf("爬取论文失败: %v", err)
	}

	// 限制数量
	if limit > 0 && len(papers) > limit {
		papers = papers[:limit]
	}

	// 2. 批量分析（带缓存）
	results, err := s.AnalyzePapersWithCache(papers)
	if err != nil {
		return nil, err
	}

	// 3. 按评分排序，返回Top N
	topPapers := GetTopScoredPapers(results, topN)

	logrus.Infof("论文分析完成，从 %d 篇中选出 %d 篇高分论文用于写作", len(papers), len(topPapers))
	return topPapers, nil
}

// AnalyzePapersFromList 从给定论文列表中分析并选择高分论文
func (s *AutogenService) AnalyzePapersFromList(papers []crawler_service.ArxivPaper, topN int) ([]*PaperAnalysisResult, error) {
	// 1. 批量分析（不缓存，确保评分标准一致）
	results, err := s.AnalyzePapers(papers)
	if err != nil {
		return nil, err
	}

	// 2. 按评分排序，返回Top N
	topPapers := GetTopScoredPapers(results, topN)

	logrus.Infof("论文分析完成，从 %d 篇中选出 %d 篇高分论文用于写作", len(papers), len(topPapers))
	return topPapers, nil
}

// getAnalysisFromCache 从Redis缓存获取AI分析结果
func (s *AutogenService) getAnalysisFromCache(arxivID string) *PaperAnalysisResult {
	if global.Redis == nil {
		return nil
	}

	cacheKey := AnalysisResultPrefix + arxivID
	resultJSON, err := global.Redis.Get(cacheKey).Result()
	if err != nil {
		// 缓存未命中，正常情况
		return nil
	}

	var result PaperAnalysisResult
	err = json.Unmarshal([]byte(resultJSON), &result)
	if err != nil {
		logrus.Errorf("反序列化分析结果失败 %s: %v", arxivID, err)
		// 删除损坏的缓存
		global.Redis.Del(cacheKey)
		return nil
	}

	logrus.Debugf("从缓存获取分析结果: %s", arxivID)
	return &result
}

// saveAnalysisToCache 保存AI分析结果到Redis缓存
func (s *AutogenService) saveAnalysisToCache(result *PaperAnalysisResult) {
	if global.Redis == nil || result == nil {
		return
	}

	cacheKey := AnalysisResultPrefix + result.ArxivID
	resultJSON, err := json.Marshal(result)
	if err != nil {
		logrus.Errorf("序列化分析结果失败 %s: %v", result.ArxivID, err)
		return
	}

	err = global.Redis.Set(cacheKey, resultJSON, AnalysisCacheExpiration).Err()
	if err != nil {
		logrus.Errorf("保存分析结果到缓存失败 %s: %v", result.ArxivID, err)
		return
	}

	logrus.Debugf("保存分析结果到缓存: %s (7天过期)", result.ArxivID)
}

// ClearAnalysisCache 清理分析结果缓存
func (s *AutogenService) ClearAnalysisCache() error {
	if global.Redis == nil {
		return fmt.Errorf("redis未连接")
	}

	pattern := AnalysisResultPrefix + "*"
	keys, err := global.Redis.Keys(pattern).Result()
	if err != nil {
		return fmt.Errorf("获取缓存键失败: %v", err)
	}

	if len(keys) == 0 {
		logrus.Info("没有需要清理的分析缓存")
		return nil
	}

	deleted, err := global.Redis.Del(keys...).Result()
	if err != nil {
		return fmt.Errorf("删除缓存失败: %v", err)
	}

	logrus.Infof("成功清理 %d 个分析结果缓存", deleted)
	return nil
}

// GetCacheStats 获取分析缓存统计信息
func (s *AutogenService) GetCacheStats() (map[string]interface{}, error) {
	if global.Redis == nil {
		return nil, fmt.Errorf("redis未连接")
	}

	pattern := AnalysisResultPrefix + "*"
	keys, err := global.Redis.Keys(pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("获取缓存键失败: %v", err)
	}

	stats := map[string]interface{}{
		"total_cached":    len(keys),
		"cache_prefix":    AnalysisResultPrefix,
		"expiration_days": int(AnalysisCacheExpiration.Hours() / 24),
	}

	if len(keys) > 0 {
		// 获取一个样本的TTL
		ttl, err := global.Redis.TTL(keys[0]).Result()
		if err == nil {
			stats["sample_ttl_hours"] = int(ttl.Hours())
		}
	}

	return stats, nil
}

// calculateAverage 计算平均值
func calculateAverage(scores []int) float64 {
	if len(scores) == 0 {
		return 0
	}

	sum := 0
	for _, score := range scores {
		sum += score
	}

	return float64(sum) / float64(len(scores))
}

// findMax 找最大值
func findMax(scores []int) int {
	if len(scores) == 0 {
		return 0
	}

	max := scores[0]
	for _, score := range scores {
		if score > max {
			max = score
		}
	}

	return max
}

// findMin 找最小值
func findMin(scores []int) int {
	if len(scores) == 0 {
		return 0
	}

	min := scores[0]
	for _, score := range scores {
		if score < min {
			min = score
		}
	}

	return min
}

// getScoreEmoji 根据评分返回表情符号
func getScoreEmoji(score int) string {
	switch {
	case score >= 90:
		return "🔥" // 优秀
	case score >= 80:
		return "⭐" // 良好
	case score >= 70:
		return "👍" // 不错
	case score >= 60:
		return "👌" // 一般
	default:
		return "📝" // 普通
	}
}
