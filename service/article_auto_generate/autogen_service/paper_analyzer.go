// Package autogen_service 提供论文自动分析服务
//
// **重要提示**: 本文件中的分析函数已被新的批次评分系统取代
// 推荐使用: service/batch_scoring_service 进行论文评分和分析
// 当前文件仅保留缓存管理功能，分析功能标记为 @deprecated
package autogen_service

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"blogX_server/global"
	"blogX_server/service/ai_service"
	"blogX_server/service/article_auto_generate/crawler_service"
	"blogX_server/service/common_utils"

	"github.com/sirupsen/logrus"
)

const (
	// AI分析结果Redis缓存键前缀
	AnalysisResultPrefix = "ai_analysis:"
	// 分析结果缓存时间（7天）
	AnalysisCacheExpiration = 23 * time.Hour
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
// @deprecated 推荐使用 batch_scoring_service.TwoStageAnalyzer 进行批次评分
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

	// 5. 清理无效的JSON转义序列
	cleanResponse = cleanInvalidJSONEscapes(cleanResponse)

	// 6. 解析AI返回的JSON结果
	var aiResult AIAnalysisResponse
	err = json.Unmarshal([]byte(cleanResponse), &aiResult)
	if err != nil {
		logrus.Errorf("AI返回结果解析失败: %v, 原始结果: %s", err, aiResponse)
		return nil, fmt.Errorf("AI返回结果解析失败: %v", err)
	}

	// 7. 构建最终结果
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

	// 8. 自动保存到缓存
	s.saveAnalysisToCache(result)

	logrus.Infof("论文 %s 分析完成，评分: %d（已缓存）", paper.ArxivID, result.Score)
	return result, nil
}

// AnalyzePapers 批量分析论文（自动缓存）
// @deprecated 推荐使用 batch_scoring_service.TwoStageAnalyzer 进行批次评分
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
// @deprecated 推荐使用 batch_scoring_service.TwoStageAnalyzer 进行批次评分
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
// @deprecated 推荐使用 article_autogen.formatTwoStageAnalysisReport
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
	avgScore := common_utils.CalculateAverage(scores)
	maxScore := common_utils.FindMax(scores)
	minScore := common_utils.FindMin(scores)

	body.WriteString(fmt.Sprintf("- **最高分**: `%d`  \n", maxScore))
	body.WriteString(fmt.Sprintf("- **平均分**: `%.1f`  \n", avgScore))
	body.WriteString(fmt.Sprintf("- **最低分**: `%d`  \n\n", minScore))

	body.WriteString("---\n\n")

	// 📝 论文详情
	for i, paper := range results {
		// 👥 作者名称截断处理
		authors := common_utils.TruncateAuthor(paper.Authors, 100)

		body.WriteString(fmt.Sprintf("### %02d %s\n\n", i, paper.Title))
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
		body.WriteString(fmt.Sprintf("**AI摘要**: %s\n\n", paper.Abstract))
		body.WriteString(fmt.Sprintf("**本站评分**: `%d/100`\n", paper.Score))
		body.WriteString(fmt.Sprintf("**评分分析**: %s  \n", paper.Justification))

		// 分隔符（最后一篇不添加）
		if i < len(results)-1 {
			body.WriteString("---\n\n")
		}
	}

	return body.String()
}

// AnalyzePapersForWriting 专门用于写文章的论文分析（推荐使用）
// @deprecated 推荐使用 batch_scoring_service.TwoStageAnalyzer 进行批次评分
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
// @deprecated 推荐使用 batch_scoring_service.TwoStageAnalyzer 进行批次评分
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
// @deprecated 使用 common_utils.CalculateAverage 替代
func calculateAverage(scores []int) float64 {
	return common_utils.CalculateAverage(scores)
}

// findMax 找最大值
// @deprecated 使用 common_utils.FindMax 替代
func findMax(scores []int) int {
	return common_utils.FindMax(scores)
}

// findMin 找最小值
// @deprecated 使用 common_utils.FindMin 替代
func findMin(scores []int) int {
	return common_utils.FindMin(scores)
}

// getScoreEmoji 根据评分返回表情符号
// @deprecated 使用 common_utils.GetScoreEmoji 替代
func getScoreEmoji(score int) string {
	return common_utils.GetScoreEmoji(float64(score))
}

// cleanInvalidJSONEscapes 清理无效的JSON转义序列
func cleanInvalidJSONEscapes(jsonStr string) string {
	// 定义有效的JSON转义字符
	validEscapes := map[string]bool{
		"\\\"": true, // 引号
		"\\\\": true, // 反斜杠
		"\\/":  true, // 斜杠
		"\\b":  true, // 退格
		"\\f":  true, // 换页
		"\\n":  true, // 换行
		"\\r":  true, // 回车
		"\\t":  true, // 制表符
	}

	var result strings.Builder
	i := 0

	for i < len(jsonStr) {
		if jsonStr[i] == '\\' && i+1 < len(jsonStr) {
			// 检查是否是Unicode转义 \uXXXX
			if jsonStr[i+1] == 'u' && i+5 < len(jsonStr) {
				// 检查后面4个字符是否都是十六进制
				isValidUnicode := true
				for j := i + 2; j < i+6; j++ {
					c := jsonStr[j]
					if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
						isValidUnicode = false
						break
					}
				}

				if isValidUnicode {
					// 有效的Unicode转义，保留
					result.WriteString(jsonStr[i : i+6])
					i += 6
					continue
				}
			}

			// 检查是否是有效的2字符转义序列
			escapeSeq := jsonStr[i : i+2]
			if validEscapes[escapeSeq] {
				// 有效转义，保留
				result.WriteString(escapeSeq)
				i += 2
			} else {
				// 无效转义，移除反斜杠
				logrus.Warnf("清理无效JSON转义序列: %s", escapeSeq)
				result.WriteByte(jsonStr[i+1]) // 只保留转义后的字符
				i += 2
			}
		} else {
			// 普通字符，直接添加
			result.WriteByte(jsonStr[i])
			i++
		}
	}

	return result.String()
}
