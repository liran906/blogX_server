package ai_api

import (
	"blogX_server/common/res"
	"blogX_server/service/batch_scoring_service"
	"blogX_server/service/crawler_service"
	"blogX_server/service/redis_service/redis_ai_cache"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// BatchAnalyzeRequest 批次分析请求
type BatchAnalyzeRequest struct {
	Papers []crawler_service.ArxivPaper `json:"papers" binding:"required"`
}

// BatchAnalyzeResponse 批次分析响应
type BatchAnalyzeResponse struct {
	Stage1Results  []batch_scoring_service.PaperScore       `json:"stage1_results"`
	Stage2Results  []batch_scoring_service.DetailedAnalysis `json:"stage2_results"`
	Statistics     batch_scoring_service.AnalysisStatistics `json:"statistics"`
	ProcessingTime string                                   `json:"processing_time"`
	Stage1Time     string                                   `json:"stage1_time"`
	Stage2Time     string                                   `json:"stage2_time"`
}

// BatchAnalyzePapers 批次分析论文
func (AiApi) BatchAnalyzePapers(c *gin.Context) {
	var req BatchAnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		res.FailWithMsg("参数绑定失败: "+err.Error(), c)
		return
	}

	if len(req.Papers) == 0 {
		res.FailWithMsg("论文列表不能为空", c)
		return
	}

	if len(req.Papers) > 1000 {
		res.FailWithMsg("论文数量不能超过1000篇", c)
		return
	}

	logrus.Infof("开始批次分析：%d篇论文", len(req.Papers))

	// 创建两阶段分析器
	config := batch_scoring_service.DefaultBatchScoringConfig()
	analyzer := batch_scoring_service.NewTwoStageAnalyzer(config)

	// 执行两阶段分析
	request := batch_scoring_service.TwoStageAnalysisRequest{
		Papers: req.Papers,
	}

	result, err := analyzer.AnalyzeTwoStage(request)
	if err != nil {
		logrus.Errorf("批次分析失败: %v", err)
		res.FailWithMsg("批次分析失败: "+err.Error(), c)
		return
	}

	// 保存结果到缓存
	go func() {
		err := redis_ai_cache.SaveBatchScoringResult(result)
		if err != nil {
			logrus.Errorf("保存缓存失败: %v", err)
		}
	}()

	// 构建响应
	response := BatchAnalyzeResponse{
		Stage1Results:  result.Stage1Results,
		Stage2Results:  result.Stage2Results,
		Statistics:     result.Statistics,
		ProcessingTime: result.ProcessingTime.String(),
		Stage1Time:     result.Stage1Time.String(),
		Stage2Time:     result.Stage2Time.String(),
	}

	logrus.Infof("批次分析完成：第一阶段%d篇，第二阶段%d篇，总耗时%v",
		len(result.Stage1Results), len(result.Stage2Results), result.ProcessingTime)

	res.SuccessWithData(response, c)
}

// QuickAnalyzeRequest 快速分析请求
type QuickAnalyzeRequest struct {
	Papers []crawler_service.ArxivPaper `json:"papers" binding:"required"`
	TopN   int                          `json:"top_n"`
}

// QuickAnalyzePapers 快速分析论文（只进行第一阶段评分）
func (AiApi) QuickAnalyzePapers(c *gin.Context) {
	var req QuickAnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		res.FailWithMsg("参数绑定失败: "+err.Error(), c)
		return
	}

	if len(req.Papers) == 0 {
		res.FailWithMsg("论文列表不能为空", c)
		return
	}

	if req.TopN <= 0 {
		req.TopN = 20 // 默认返回前20篇
	}

	logrus.Infof("开始快速分析：%d篇论文，返回前%d篇", len(req.Papers), req.TopN)

	// 创建配置，只进行第一阶段
	config := batch_scoring_service.DefaultBatchScoringConfig()
	config.TopN = 0 // 设置为0表示不进行第二阶段

	analyzer := batch_scoring_service.NewTwoStageAnalyzer(config)

	// 执行第一阶段分析
	request := batch_scoring_service.TwoStageAnalysisRequest{
		Papers: req.Papers,
	}

	result, err := analyzer.AnalyzeTwoStage(request)
	if err != nil {
		logrus.Errorf("快速分析失败: %v", err)
		res.FailWithMsg("快速分析失败: "+err.Error(), c)
		return
	}

	// 只返回前TopN篇论文的评分
	stage1Results := result.Stage1Results
	if len(stage1Results) > req.TopN {
		// 按分数排序并截取前TopN
		sort.Slice(stage1Results, func(i, j int) bool {
			return stage1Results[i].FinalScore > stage1Results[j].FinalScore
		})
		stage1Results = stage1Results[:req.TopN]
	}

	response := map[string]interface{}{
		"results":         stage1Results,
		"statistics":      result.Statistics,
		"processing_time": result.ProcessingTime.String(),
		"total_papers":    len(req.Papers),
		"returned_papers": len(stage1Results),
	}

	logrus.Infof("快速分析完成：评分%d篇论文，返回前%d篇，耗时%v",
		len(result.Stage1Results), len(stage1Results), result.ProcessingTime)

	res.SuccessWithData(response, c)
}

// CheckCacheRequest 缓存检查请求
type CheckCacheRequest struct {
	Papers []struct {
		ArxivID  string `json:"arxiv_id" binding:"required"`
		Title    string `json:"title" binding:"required"`
		Abstract string `json:"abstract" binding:"required"`
	} `json:"papers" binding:"required"`
}

// CheckPaperCache 检查论文缓存状态
func (AiApi) CheckPaperCache(c *gin.Context) {
	var req CheckCacheRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		res.FailWithMsg("参数绑定失败: "+err.Error(), c)
		return
	}

	results := make([]map[string]interface{}, len(req.Papers))

	for i, paper := range req.Papers {
		// 检查详细分析缓存
		hasDetailedCache := redis_ai_cache.CheckCacheExists(
			paper.ArxivID, paper.Title, paper.Abstract, "detailed")

		// 检查评分缓存
		hasScoringCache := redis_ai_cache.CheckCacheExists(
			paper.ArxivID, paper.Title, paper.Abstract, "scoring")

		results[i] = map[string]interface{}{
			"arxiv_id":           paper.ArxivID,
			"has_detailed_cache": hasDetailedCache,
			"has_scoring_cache":  hasScoringCache,
			"cache_available":    hasDetailedCache || hasScoringCache,
		}
	}

	response := map[string]interface{}{
		"results":      results,
		"total_papers": len(req.Papers),
		"cached_count": func() int {
			count := 0
			for _, result := range results {
				if result["cache_available"].(bool) {
					count++
				}
			}
			return count
		}(),
	}

	res.SuccessWithData(response, c)
}

// GetCachedAnalysis 获取缓存的分析结果
func (AiApi) GetCachedAnalysis(c *gin.Context) {
	paperID := c.Param("paper_id")
	if paperID == "" {
		res.FailWithMsg("论文ID不能为空", c)
		return
	}

	title := c.Query("title")
	abstract := c.Query("abstract")

	if title == "" || abstract == "" {
		res.FailWithMsg("标题和摘要参数不能为空", c)
		return
	}

	// 尝试获取详细分析缓存
	detailedAnalysis, err := redis_ai_cache.GetDetailedAnalysis(paperID, title, abstract)
	if err == nil {
		logrus.Infof("返回缓存的详细分析：%s", paperID)
		res.SuccessWithData(map[string]interface{}{
			"type":     "detailed",
			"analysis": detailedAnalysis,
		}, c)
		return
	}

	// 尝试获取评分缓存
	score, reasoning, err := redis_ai_cache.GetPaperScore(paperID, title, abstract)
	if err == nil {
		logrus.Infof("返回缓存的评分结果：%s = %.1f", paperID, score)
		res.SuccessWithData(map[string]interface{}{
			"type":      "scoring",
			"score":     score,
			"reasoning": reasoning,
		}, c)
		return
	}

	res.FailWithMsg("未找到缓存的分析结果", c)
}

// GetCacheStats 获取缓存统计信息
func (AiApi) GetCacheStats(c *gin.Context) {
	stats := redis_ai_cache.GetCacheStats()
	res.SuccessWithData(stats, c)
}

// ClearPaperCache 清除指定论文的缓存
func (AiApi) ClearPaperCache(c *gin.Context) {
	paperID := c.Param("paper_id")
	if paperID == "" {
		res.FailWithMsg("论文ID不能为空", c)
		return
	}

	err := redis_ai_cache.InvalidateCache(paperID)
	if err != nil {
		logrus.Errorf("清除缓存失败: %v", err)
		res.FailWithMsg("清除缓存失败: "+err.Error(), c)
		return
	}

	logrus.Infof("已清除论文 %s 的缓存", paperID)
	res.SuccessWithMsg("缓存清除成功", c)
}

// CrawlAndAnalyzeRequest 爬取并分析请求
type CrawlAndAnalyzeRequest struct {
	Category  string `json:"category" binding:"required"`
	TopN      int    `json:"top_n"`
	MaxPapers int    `json:"max_papers"`
}

// CrawlAndAnalyzePapers 爬取论文并进行批次分析
func (AiApi) CrawlAndAnalyzePapers(c *gin.Context) {
	var req CrawlAndAnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		res.FailWithMsg("参数绑定失败: "+err.Error(), c)
		return
	}

	if req.TopN <= 0 {
		req.TopN = 20
	}
	if req.MaxPapers <= 0 {
		req.MaxPapers = 200
	}

	logrus.Infof("开始爬取并分析：类别%s，最多%d篇论文，返回前%d篇",
		req.Category, req.MaxPapers, req.TopN)

	// 根据类别获取爬虫
	var papers []crawler_service.ArxivPaper
	var err error

	switch req.Category {
	case "ai", "AI":
		papers, err = crawler_service.CrawlPapersByCategory(crawler_service.CategoryAI)
	case "astro", "astrophysics":
		papers, err = crawler_service.CrawlAstrophysicsPapers()
	case "cs", "computer_science":
		papers, err = crawler_service.CrawlComputerSciencePapers()
	case "math", "mathematics":
		papers, err = crawler_service.CrawlMathematicsPapers()
	case "physics":
		papers, err = crawler_service.CrawlPhysicsPapers()
	case "quantum":
		papers, err = crawler_service.CrawlQuantumPhysicsPapers()
	default:
		res.FailWithMsg("不支持的论文类别: "+req.Category, c)
		return
	}

	if err != nil {
		logrus.Errorf("爬取论文失败: %v", err)
		res.FailWithMsg("爬取论文失败: "+err.Error(), c)
		return
	}

	if len(papers) == 0 {
		res.FailWithMsg("未爬取到论文", c)
		return
	}

	// 限制论文数量
	if len(papers) > req.MaxPapers {
		papers = papers[:req.MaxPapers]
	}

	logrus.Infof("成功爬取 %d 篇论文，开始分析", len(papers))

	// 执行批次分析
	config := batch_scoring_service.DefaultBatchScoringConfig()
	config.TopN = req.TopN

	analyzer := batch_scoring_service.NewTwoStageAnalyzer(config)

	request := batch_scoring_service.TwoStageAnalysisRequest{
		Papers: papers,
	}

	result, err := analyzer.AnalyzeTwoStage(request)
	if err != nil {
		logrus.Errorf("论文分析失败: %v", err)
		res.FailWithMsg("论文分析失败: "+err.Error(), c)
		return
	}

	// 异步保存到缓存
	go func() {
		err := redis_ai_cache.SaveBatchScoringResult(result)
		if err != nil {
			logrus.Errorf("保存缓存失败: %v", err)
		}
	}()

	response := map[string]interface{}{
		"category":         req.Category,
		"crawled_papers":   len(papers),
		"stage1_results":   result.Stage1Results,
		"stage2_results":   result.Stage2Results,
		"statistics":       result.Statistics,
		"processing_time":  result.ProcessingTime.String(),
		"average_score":    result.Statistics.AverageScore,
		"top_papers_count": len(result.Stage2Results),
	}

	logrus.Infof("爬取并分析完成：%s类别，%d篇论文，平均分%.1f，前%d篇详细分析",
		req.Category, len(papers), result.Statistics.AverageScore, len(result.Stage2Results))

	res.SuccessWithData(response, c)
}

// GetAnalysisConfig 获取分析配置
func (AiApi) GetAnalysisConfig(c *gin.Context) {
	config := batch_scoring_service.DefaultBatchScoringConfig()

	response := map[string]interface{}{
		"batch_size":           config.BatchSize,
		"score_diff_threshold": config.ScoreDiffThreshold,
		"max_retries":          config.MaxRetries,
		"top_n":                config.TopN,
		"cache_expiry_days":    7,
		"supported_categories": []string{"ai", "astro", "cs", "math", "physics", "quantum"},
	}

	res.SuccessWithData(response, c)
}

// UpdateAnalysisConfig 更新分析配置（管理员功能）
func (AiApi) UpdateAnalysisConfig(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		res.FailWithMsg("参数绑定失败: "+err.Error(), c)
		return
	}

	// 这里可以根据需要更新全局配置
	// 目前只是返回确认信息
	logrus.Infof("配置更新请求: %+v", req)

	res.SuccessWithMsg("配置更新成功（模拟）", c)
}
