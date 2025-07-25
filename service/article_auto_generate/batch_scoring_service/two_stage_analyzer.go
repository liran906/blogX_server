package batch_scoring_service

import (
	"blogX_server/service/ai_service"
	"blogX_server/service/article_auto_generate/crawler_service"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// TwoStageAnalyzer 两阶段分析器
type TwoStageAnalyzer struct {
	config         *BatchScoringConfig
	batchAllocator *BatchAllocator
	batchScorer    *BatchScorer
}

// NewTwoStageAnalyzer 创建两阶段分析器
func NewTwoStageAnalyzer(config *BatchScoringConfig) *TwoStageAnalyzer {
	return &TwoStageAnalyzer{
		config:         config,
		batchAllocator: NewBatchAllocator(config),
		batchScorer:    NewBatchScorer(config),
	}
}

// TwoStageAnalysisRequest 两阶段分析请求
type TwoStageAnalysisRequest struct {
	Papers []crawler_service.ArxivPaper // 待分析的论文列表
}

// TwoStageAnalysisResult 两阶段分析结果
type TwoStageAnalysisResult struct {
	Stage1Results  []PaperScore       // 第一阶段批次评分结果
	Stage2Results  []DetailedAnalysis // 第二阶段详细分析结果
	Statistics     AnalysisStatistics // 统计信息
	ProcessingTime time.Duration      // 总处理时间
	Stage1Time     time.Duration      // 第一阶段耗时
	Stage2Time     time.Duration      // 第二阶段耗时
}

// DetailedAnalysis 详细分析结果
type DetailedAnalysis struct {
	ArxivID    string   `json:"arxiv_id"`
	Title      string   `json:"title"`
	Authors    string   `json:"authors"`
	Abstract   string   `json:"abstract"`
	Tags       []string `json:"tags"`
	Evaluation string   `json:"evaluation"` // 专业评价
	Summary    string   `json:"summary"`    // 中文摘要
}

// AnalysisStatistics 分析统计信息
type AnalysisStatistics struct {
	TotalPapers         int            // 总论文数
	Stage1Batches       int            // 第一阶段批次数
	ConflictPapers      int            // 冲突论文数
	ThirdRoundPapers    int            // 需要第三次评分的论文数
	Stage2SelectedCount int            // 第二阶段选中论文数
	BatchRetries        map[int]int    // 批次重试次数统计
	ScoreDistribution   map[string]int // 分数分布统计
	AverageScore        float64        // 平均分数
	MaxScore            float64        // 最高分数
	MinScore            float64        // 最低分数
}

// AnalyzeTwoStage 执行两阶段分析
func (tsa *TwoStageAnalyzer) AnalyzeTwoStage(request TwoStageAnalysisRequest) (*TwoStageAnalysisResult, error) {
	startTime := time.Now()

	logrus.Infof("开始两阶段分析：论文数量=%d", len(request.Papers))

	// 第一阶段：批次评分
	stage1Start := time.Now()
	stage1Results, stage1Stats, err := tsa.runStage1BatchScoring(request.Papers)
	if err != nil {
		return nil, fmt.Errorf("第一阶段批次评分失败: %v", err)
	}
	stage1Duration := time.Since(stage1Start)

	logrus.Infof("第一阶段完成：耗时%v，评分%d篇论文", stage1Duration, len(stage1Results))

	// 第二阶段：详细分析
	stage2Start := time.Now()
	stage2Results, err := tsa.runStage2DetailedAnalysis(stage1Results, request.Papers)
	if err != nil {
		return nil, fmt.Errorf("第二阶段详细分析失败: %v", err)
	}
	stage2Duration := time.Since(stage2Start)

	logrus.Infof("第二阶段完成：耗时%v，详细分析%d篇论文", stage2Duration, len(stage2Results))

	// 计算总体统计信息
	totalDuration := time.Since(startTime)
	statistics := tsa.calculateStatistics(stage1Results, stage1Stats, len(stage2Results))

	result := &TwoStageAnalysisResult{
		Stage1Results:  stage1Results,
		Stage2Results:  stage2Results,
		Statistics:     statistics,
		ProcessingTime: totalDuration,
		Stage1Time:     stage1Duration,
		Stage2Time:     stage2Duration,
	}

	logrus.Infof("两阶段分析完成：总耗时%v，平均分%.1f，选中%d篇论文进行详细分析",
		totalDuration, statistics.AverageScore, len(stage2Results))

	return result, nil
}

// runStage1BatchScoring 执行第一阶段批次评分
func (tsa *TwoStageAnalyzer) runStage1BatchScoring(papers []crawler_service.ArxivPaper) ([]PaperScore, map[int]int, error) {
	logrus.Infof("开始第一阶段：批次评分，论文数量=%d", len(papers))

	// 1. 分配论文到批次
	allocation, err := tsa.batchAllocator.AllocatePapersToBatches(len(papers))
	if err != nil {
		return nil, nil, fmt.Errorf("批次分配失败: %v", err)
	}

	// 2. 并行执行批次评分
	batchResults, retryStats, err := tsa.executeBatchScoring(allocation, papers)
	if err != nil {
		return nil, nil, fmt.Errorf("批次评分执行失败: %v", err)
	}

	// 3. 合并批次结果
	paperScores, err := tsa.mergeBatchResults(batchResults, papers, allocation)
	if err != nil {
		return nil, nil, fmt.Errorf("批次结果合并失败: %v", err)
	}

	logrus.Infof("第一阶段批次评分完成：%d个批次，%d篇论文", allocation.TotalBatches, len(paperScores))

	return paperScores, retryStats, nil
}

// executeBatchScoring 并行执行批次评分
func (tsa *TwoStageAnalyzer) executeBatchScoring(allocation *BatchAllocation, papers []crawler_service.ArxivPaper) (map[int]*BatchScoringResponse, map[int]int, error) {
	batchCount := allocation.TotalBatches
	results := make(map[int]*BatchScoringResponse)
	retryStats := make(map[int]int)

	// 使用goroutine并行处理批次
	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := make([]error, 0)

	for batchID := 0; batchID < batchCount; batchID++ {
		wg.Add(1)
		go func(bid int) {
			defer wg.Done()

			// 准备这个批次的论文
			batchPapers := make([]crawler_service.ArxivPaper, 0, tsa.config.BatchSize)
			for _, paperIndex := range allocation.Batches[bid] {
				if paperIndex < len(papers) {
					batchPapers = append(batchPapers, papers[paperIndex])
				}
			}

			// 执行批次评分（包含重试机制）
			result, retries := tsa.scoreBatchWithRetry(bid, batchPapers)

			mu.Lock()
			if result.Success {
				results[bid] = result
				retryStats[bid] = retries
			} else {
				errors = append(errors, fmt.Errorf("批次 %d 最终失败: %s", bid, result.Error))
			}
			mu.Unlock()
		}(batchID)
	}

	wg.Wait()

	if len(errors) > 0 {
		return nil, nil, fmt.Errorf("有 %d 个批次失败: %v", len(errors), errors[0])
	}

	return results, retryStats, nil
}

// scoreBatchWithRetry 带重试机制的批次评分
func (tsa *TwoStageAnalyzer) scoreBatchWithRetry(batchID int, papers []crawler_service.ArxivPaper) (*BatchScoringResponse, int) {
	maxRetries := tsa.config.MaxRetries

	for attempt := 0; attempt <= maxRetries; attempt++ {
		request := BatchScoringRequest{
			BatchID: batchID,
			Papers:  papers,
			Attempt: attempt + 1,
		}

		result, err := tsa.batchScorer.ScoreBatch(request)
		if err == nil && result.Success {
			logrus.Infof("批次 %d 评分成功，尝试次数: %d", batchID, attempt+1)
			return result, attempt
		}

		if attempt < maxRetries {
			logrus.Warnf("批次 %d 第 %d 次尝试失败，重试中: %v", batchID, attempt+1, err)
			time.Sleep(time.Second * 2) // 等待2秒后重试
		} else {
			logrus.Errorf("批次 %d 达到最大重试次数 %d，最终失败", batchID, maxRetries+1)
			return &BatchScoringResponse{
				BatchID: batchID,
				Success: false,
				Error:   fmt.Sprintf("达到最大重试次数: %v", err),
			}, attempt
		}
	}

	return nil, maxRetries
}

// mergeBatchResults 合并批次结果，处理冲突检测和第三次评分
func (tsa *TwoStageAnalyzer) mergeBatchResults(batchResults map[int]*BatchScoringResponse, papers []crawler_service.ArxivPaper, allocation *BatchAllocation) ([]PaperScore, error) {
	paperScores := make([]PaperScore, 0, len(papers))
	conflictPapers := make([]crawler_service.ArxivPaper, 0)
	paperMap := make(map[string]crawler_service.ArxivPaper)

	// 建立论文映射
	for _, paper := range papers {
		paperMap[paper.ArxivID] = paper
	}

	// 第一步：收集所有论文的前两次评分，并识别冲突论文
	paperFirstTwoScores := make(map[string]*struct {
		Score1   *DetailedScore
		Score2   *DetailedScore
		BatchIDs []int
		Paper    crawler_service.ArxivPaper
	})

	for paperIndex, paper := range papers {
		batches := allocation.PaperToBatches[paperIndex]
		if len(batches) != 2 {
			return nil, fmt.Errorf("论文 %s 分配的批次数量异常: %d", paper.ArxivID, len(batches))
		}

		// 获取两个批次的评分
		score1, score2, err := tsa.getBatchScoresForPaper(paper.ArxivID, batches, batchResults)
		if err != nil {

			logrus.Warnf("获取论文 %s 的批次评分失败: %v, 丢弃本篇论文", paper.ArxivID, err)

			continue
		}

		paperFirstTwoScores[paper.ArxivID] = &struct {
			Score1   *DetailedScore
			Score2   *DetailedScore
			BatchIDs []int
			Paper    crawler_service.ArxivPaper
		}{
			Score1:   score1,
			Score2:   score2,
			BatchIDs: batches,
			Paper:    paper,
		}

		// 检测冲突
		if tsa.batchScorer.DetectScoreConflict(score1, score2) {
			logrus.Warnf("论文 %s 分数冲突：总分 %d vs %d，需要第三次评分",
				paper.ArxivID, score1.Total, score2.Total)
			conflictPapers = append(conflictPapers, paper)
		}
	}

	// 第二步：对冲突论文进行第三次批次评分
	var thirdRoundScores map[string]*DetailedScore
	if len(conflictPapers) > 0 {
		logrus.Infof("开始第三次批次评分，共%d篇冲突论文", len(conflictPapers))

		// 按batch size分组处理第三次评分
		thirdRoundScores = make(map[string]*DetailedScore)
		batchSize := tsa.config.ThirdRoundBatchSize

		for i := 0; i < len(conflictPapers); i += batchSize {
			end := i + batchSize
			if end > len(conflictPapers) {
				end = len(conflictPapers)
			}

			subBatch := conflictPapers[i:end]
			scores, err := tsa.batchScorer.ScoreThirdRoundBatch(subBatch)
			if err != nil {
				logrus.Errorf("第三次批次评分失败: %v", err)
				// 对于失败的论文，设置为nil
				for _, paper := range subBatch {
					thirdRoundScores[paper.ArxivID] = nil
				}
			} else {
				// 合并结果
				for arxivID, score := range scores {
					thirdRoundScores[arxivID] = score
				}
			}
		}
	}

	// 第三步：合并所有评分结果
	for _, paperData := range paperFirstTwoScores {
		var score3 *DetailedScore
		status := StatusCompleted

		// 检查是否需要第三次评分
		if _, hasConflict := thirdRoundScores[paperData.Paper.ArxivID]; hasConflict {
			score3 = thirdRoundScores[paperData.Paper.ArxivID]
			if score3 != nil {
				status = StatusThirdRound
			} else {
				status = StatusFailed
			}
		}

		// 计算最终分数
		finalScore := tsa.batchScorer.MergeFinalScore(paperData.Score1, paperData.Score2, score3)

		paperScores = append(paperScores, PaperScore{
			ArxivID:    paperData.Paper.ArxivID,
			Score1:     paperData.Score1,
			Score2:     paperData.Score2,
			Score3:     score3,
			FinalScore: finalScore,
			BatchIDs:   paperData.BatchIDs,
			Status:     status,
		})
	}

	return paperScores, nil
}

// getBatchScoresForPaper 获取论文在两个批次中的评分
func (tsa *TwoStageAnalyzer) getBatchScoresForPaper(arxivID string, batchIDs []int, batchResults map[int]*BatchScoringResponse) (*DetailedScore, *DetailedScore, error) {
	if len(batchIDs) != 2 {
		return nil, nil, fmt.Errorf("论文 %s 的批次数量异常: %d", arxivID, len(batchIDs))
	}

	batch1Result, exists1 := batchResults[batchIDs[0]]
	if !exists1 || !batch1Result.Success {
		return nil, nil, fmt.Errorf("批次 %d 结果不存在或失败", batchIDs[0])
	}

	batch2Result, exists2 := batchResults[batchIDs[1]]
	if !exists2 || !batch2Result.Success {
		return nil, nil, fmt.Errorf("批次 %d 结果不存在或失败", batchIDs[1])
	}

	// 在批次结果中查找论文评分
	var score1, score2 *DetailedScore
	var found1, found2 bool

	for _, result := range batch1Result.Results {
		if result.ArxivID == arxivID {
			score1 = &DetailedScore{
				Innovation: result.Innovation,
				Technical:  result.Technical,
				Practical:  result.Practical,
				Total:      result.Total,
			}
			found1 = true
			break
		}
	}

	for _, result := range batch2Result.Results {
		if result.ArxivID == arxivID {
			score2 = &DetailedScore{
				Innovation: result.Innovation,
				Technical:  result.Technical,
				Practical:  result.Practical,
				Total:      result.Total,
			}
			found2 = true
			break
		}
	}

	if !found1 || !found2 {
		return nil, nil, fmt.Errorf("论文 %s 在批次结果中未找到: batch1=%v, batch2=%v", arxivID, found1, found2)
	}

	return score1, score2, nil
}

// runStage2DetailedAnalysis 执行第二阶段详细分析
func (tsa *TwoStageAnalyzer) runStage2DetailedAnalysis(stage1Results []PaperScore, papers []crawler_service.ArxivPaper) ([]DetailedAnalysis, error) {
	logrus.Infof("开始第二阶段：详细分析，选择前%d篇高分论文", tsa.config.TopN)

	// 1. 按分数排序，选择TopN
	sortedResults := make([]PaperScore, len(stage1Results))
	copy(sortedResults, stage1Results)

	sort.Slice(sortedResults, func(i, j int) bool {
		return sortedResults[i].FinalScore > sortedResults[j].FinalScore
	})

	// 选择TopN论文
	topN := tsa.config.TopN
	if topN > len(sortedResults) {
		topN = len(sortedResults)
	}

	selectedPapers := sortedResults[:topN]

	// 2. 建立论文映射
	paperMap := make(map[string]crawler_service.ArxivPaper)
	for _, paper := range papers {
		paperMap[paper.ArxivID] = paper
	}

	// 3. 并行执行详细分析
	detailedResults := make([]DetailedAnalysis, 0, topN)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, paperScore := range selectedPapers {
		wg.Add(1)
		go func(ps PaperScore) {
			defer wg.Done()

			paper, exists := paperMap[ps.ArxivID]
			if !exists {
				logrus.Errorf("论文 %s 在原始列表中未找到", ps.ArxivID)
				return
			}

			analysis, err := tsa.analyzeIndividualPaper(paper)
			if err != nil {
				logrus.Errorf("论文 %s 详细分析失败: %v", paper.ArxivID, err)
				return
			}

			mu.Lock()
			detailedResults = append(detailedResults, *analysis)
			mu.Unlock()
		}(paperScore)
	}

	wg.Wait()

	logrus.Infof("第二阶段详细分析完成：成功分析%d篇论文", len(detailedResults))

	return detailedResults, nil
}

// analyzeIndividualPaper 对单篇论文进行详细分析
func (tsa *TwoStageAnalyzer) analyzeIndividualPaper(paper crawler_service.ArxivPaper) (*DetailedAnalysis, error) {
	logrus.Debugf("开始详细分析论文: %s", paper.ArxivID)

	// 构建详细分析的输入
	inputText := fmt.Sprintf("标题：%s\n内容：%s", paper.Title, paper.Abstract)

	// 调用AI进行详细分析
	rawResponse, err := ai_service.Autogen(inputText)
	if err != nil {
		return nil, fmt.Errorf("AI调用失败: %v", err)
	}

	// 解析详细分析结果
	analysis, err := tsa.parseDetailedAnalysis(rawResponse, paper)
	if err != nil {
		return nil, fmt.Errorf("解析详细分析失败: %v", err)
	}

	return analysis, nil
}

// parseDetailedAnalysis 解析详细分析结果
func (tsa *TwoStageAnalyzer) parseDetailedAnalysis(rawResponse string, paper crawler_service.ArxivPaper) (*DetailedAnalysis, error) {
	// 提取JSON部分
	jsonStart := strings.Index(rawResponse, "{")
	jsonEnd := strings.LastIndex(rawResponse, "}") + 1

	if jsonStart == -1 || jsonEnd == 0 || jsonStart >= jsonEnd {
		return nil, fmt.Errorf("响应中未找到有效的JSON: %s", rawResponse)
	}

	jsonStr := rawResponse[jsonStart:jsonEnd]

	// 解析autogen的JSON格式（新格式：只有摘要、评价、标签）
	var response struct {
		Abstract   string   `json:"abstract"`
		Evaluation string   `json:"evaluation"`
		Tags       []string `json:"tags"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &response); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %v, 原始数据: %s", err, jsonStr)
	}

	return &DetailedAnalysis{
		ArxivID:    paper.ArxivID,
		Title:      paper.Title,
		Authors:    paper.Authors,
		Abstract:   paper.Abstract,
		Tags:       response.Tags,
		Evaluation: response.Evaluation,
		Summary:    response.Abstract,
	}, nil
}

// calculateStatistics 计算分析统计信息
func (tsa *TwoStageAnalyzer) calculateStatistics(stage1Results []PaperScore, retryStats map[int]int, stage2Count int) AnalysisStatistics {
	stats := AnalysisStatistics{
		TotalPapers:         len(stage1Results),
		Stage1Batches:       len(retryStats),
		Stage2SelectedCount: stage2Count,
		BatchRetries:        retryStats,
		ScoreDistribution:   make(map[string]int),
	}

	// 计算分数统计
	var totalScore float64
	var maxScore, minScore float64
	var conflictCount, thirdRoundCount int

	for i, result := range stage1Results {
		score := result.FinalScore
		totalScore += score

		if i == 0 {
			maxScore = score
			minScore = score
		} else {
			if score > maxScore {
				maxScore = score
			}
			if score < minScore {
				minScore = score
			}
		}

		// 分数分布统计
		scoreRange := fmt.Sprintf("%d-%d", int(score/10)*10, int(score/10)*10+9)
		stats.ScoreDistribution[scoreRange]++

		// 冲突和第三次评分统计
		if result.Status == StatusThirdRound {
			conflictCount++
			thirdRoundCount++
		}
	}

	stats.ConflictPapers = conflictCount
	stats.ThirdRoundPapers = thirdRoundCount
	stats.AverageScore = totalScore / float64(len(stage1Results))
	stats.MaxScore = maxScore
	stats.MinScore = minScore

	return stats
}
