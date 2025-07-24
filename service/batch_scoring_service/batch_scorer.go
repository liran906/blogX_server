package batch_scoring_service

import (
	"blogX_server/service/ai_service"
	"blogX_server/service/crawler_service"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// BatchScorer 批次评分器
type BatchScorer struct {
	config *BatchScoringConfig
}

// NewBatchScorer 创建批次评分器
func NewBatchScorer(config *BatchScoringConfig) *BatchScorer {
	return &BatchScorer{
		config: config,
	}
}

// BatchScoringRequest 批次评分请求
type BatchScoringRequest struct {
	BatchID int                          // 批次ID
	Papers  []crawler_service.ArxivPaper // 论文列表
	Attempt int                          // 当前尝试次数
}

// BatchScoringResponse 批次评分响应
type BatchScoringResponse struct {
	BatchID  int               // 批次ID
	Results  []PaperBatchScore // 评分结果
	Success  bool              // 是否成功
	Error    string            // 错误信息
	Duration time.Duration     // 处理时长
}

// PaperBatchScore 论文批次评分结果
type PaperBatchScore struct {
	ArxivID   string // ArXiv ID
	Score     int    // 分数
	Reasoning string // 评分理由
}

// AI返回的JSON结构
type batchScoringAIResponse struct {
	Papers []struct {
		PaperID   string `json:"paper_id"`
		Score     int    `json:"score"`
		Reasoning string `json:"reasoning"`
	} `json:"papers"`
}

// ScoreBatch 对一个批次的论文进行评分
func (bs *BatchScorer) ScoreBatch(request BatchScoringRequest) (*BatchScoringResponse, error) {
	startTime := time.Now()

	logrus.Infof("开始批次评分：BatchID=%d, 论文数量=%d, 尝试次数=%d",
		request.BatchID, len(request.Papers), request.Attempt)

	// 1. 构建批次评分的输入文本
	inputText, err := bs.buildBatchInput(request.Papers)
	if err != nil {
		return &BatchScoringResponse{
			BatchID:  request.BatchID,
			Success:  false,
			Error:    fmt.Sprintf("构建输入失败: %v", err),
			Duration: time.Since(startTime),
		}, err
	}

	// 2. 调用AI进行批次评分
	rawResponse, err := ai_service.BatchScoring(inputText)
	if err != nil {
		return &BatchScoringResponse{
			BatchID:  request.BatchID,
			Success:  false,
			Error:    fmt.Sprintf("AI调用失败: %v", err),
			Duration: time.Since(startTime),
		}, err
	}

	logrus.Debugf("BatchID=%d AI原始响应: %s", request.BatchID, rawResponse)

	// 3. 解析AI响应
	results, err := bs.parseAIResponse(rawResponse, request.Papers)
	if err != nil {
		return &BatchScoringResponse{
			BatchID:  request.BatchID,
			Success:  false,
			Error:    fmt.Sprintf("解析AI响应失败: %v", err),
			Duration: time.Since(startTime),
		}, err
	}

	// 4. 验证评分结果
	if err := bs.validateBatchResults(results, request.Papers); err != nil {
		return &BatchScoringResponse{
			BatchID:  request.BatchID,
			Success:  false,
			Error:    fmt.Sprintf("评分结果验证失败: %v", err),
			Duration: time.Since(startTime),
		}, err
	}

	logrus.Infof("批次评分完成：BatchID=%d, 成功评分%d篇论文, 耗时%v",
		request.BatchID, len(results), time.Since(startTime))

	return &BatchScoringResponse{
		BatchID:  request.BatchID,
		Results:  results,
		Success:  true,
		Duration: time.Since(startTime),
	}, nil
}

// buildBatchInput 构建批次评分的输入文本
func (bs *BatchScorer) buildBatchInput(papers []crawler_service.ArxivPaper) (string, error) {
	if len(papers) == 0 {
		return "", fmt.Errorf("论文列表为空")
	}

	var builder strings.Builder

	for i, paper := range papers {
		builder.WriteString(fmt.Sprintf("【论文 %d】\n", i+1))
		builder.WriteString(fmt.Sprintf("ID: %s\n", paper.ArxivID))
		builder.WriteString(fmt.Sprintf("标题: %s\n", paper.Title))
		builder.WriteString(fmt.Sprintf("摘要: %s\n", paper.Abstract))

		// 添加分隔线
		if i < len(papers)-1 {
			builder.WriteString("\n---\n\n")
		}
	}

	return builder.String(), nil
}

// parseAIResponse 解析AI响应
func (bs *BatchScorer) parseAIResponse(rawResponse string, papers []crawler_service.ArxivPaper) ([]PaperBatchScore, error) {
	// 清理响应文本，提取JSON部分
	jsonStart := strings.Index(rawResponse, "{")
	jsonEnd := strings.LastIndex(rawResponse, "}") + 1

	if jsonStart == -1 || jsonEnd == 0 || jsonStart >= jsonEnd {
		return nil, fmt.Errorf("AI响应中未找到有效的JSON: %s", rawResponse)
	}

	jsonStr := rawResponse[jsonStart:jsonEnd]
	logrus.Debugf("提取的JSON: %s", jsonStr)

	// 解析JSON
	var aiResponse batchScoringAIResponse
	if err := json.Unmarshal([]byte(jsonStr), &aiResponse); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %v, 原始数据: %s", err, jsonStr)
	}

	// 转换为内部结构
	results := make([]PaperBatchScore, 0, len(papers))
	paperIDMap := make(map[string]string) // paper_id -> arxiv_id 的映射

	// 建立映射关系（支持序号和ArxivID两种方式）
	for i, paper := range papers {
		paperIDMap[strconv.Itoa(i+1)] = paper.ArxivID // 序号映射
		paperIDMap[paper.ArxivID] = paper.ArxivID     // ArxivID映射
	}

	for _, aiResult := range aiResponse.Papers {
		arxivID, exists := paperIDMap[aiResult.PaperID]
		if !exists {
			logrus.Warnf("AI返回了未知的论文ID: %s", aiResult.PaperID)
			continue
		}

		results = append(results, PaperBatchScore{
			ArxivID:   arxivID,
			Score:     aiResult.Score,
			Reasoning: aiResult.Reasoning,
		})
	}

	return results, nil
}

// validateBatchResults 验证批次评分结果
func (bs *BatchScorer) validateBatchResults(results []PaperBatchScore, papers []crawler_service.ArxivPaper) error {
	if len(results) != len(papers) {
		return fmt.Errorf("评分结果数量(%d)与论文数量(%d)不匹配", len(results), len(papers))
	}

	// 检查分数范围
	for _, result := range results {
		if result.Score < 0 || result.Score > 100 {
			return fmt.Errorf("论文 %s 的分数 %d 超出范围[0,100]", result.ArxivID, result.Score)
		}
	}

	// 检查分数区分度
	scores := make([]int, len(results))
	for i, result := range results {
		scores[i] = result.Score
	}
	sort.Ints(scores)

	maxScore := scores[len(scores)-1]
	minScore := scores[0]
	scoreDiff := maxScore - minScore

	if scoreDiff < 20 {
		logrus.Warnf("批次评分区分度不足：最高分%d，最低分%d，差距%d", maxScore, minScore, scoreDiff)
	} else {
		logrus.Infof("批次评分区分度良好：最高分%d，最低分%d，差距%d", maxScore, minScore, scoreDiff)
	}

	return nil
}

// DetectScoreConflict 检测两个分数是否存在冲突
func (bs *BatchScorer) DetectScoreConflict(score1, score2 int) bool {
	diff := int(math.Abs(float64(score1 - score2)))
	return diff > bs.config.ScoreDiffThreshold
}

// ScoreIndividualPaper 对单篇论文进行第三次评分
func (bs *BatchScorer) ScoreIndividualPaper(paper crawler_service.ArxivPaper) (*int, error) {
	logrus.Infof("开始第三次评分：论文 %s", paper.ArxivID)

	// 构建单篇论文的输入
	inputText := fmt.Sprintf("标题：%s\n内容：%s", paper.Title, paper.Abstract)

	// 调用AI进行单独评分
	rawResponse, err := ai_service.Autogen(inputText)
	if err != nil {
		return nil, fmt.Errorf("AI调用失败: %v", err)
	}

	// 解析单独评分的响应（使用autogen的JSON格式）
	score, err := bs.parseIndividualScore(rawResponse)
	if err != nil {
		return nil, fmt.Errorf("解析单独评分失败: %v", err)
	}

	logrus.Infof("第三次评分完成：论文 %s, 分数 %d", paper.ArxivID, *score)
	return score, nil
}

// parseIndividualScore 解析单独评分的响应
func (bs *BatchScorer) parseIndividualScore(rawResponse string) (*int, error) {
	// 清理响应文本，提取JSON部分
	jsonStart := strings.Index(rawResponse, "{")
	jsonEnd := strings.LastIndex(rawResponse, "}") + 1

	if jsonStart == -1 || jsonEnd == 0 || jsonStart >= jsonEnd {
		return nil, fmt.Errorf("响应中未找到有效的JSON: %s", rawResponse)
	}

	jsonStr := rawResponse[jsonStart:jsonEnd]

	// 解析autogen的JSON格式
	var response struct {
		Score int `json:"score"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &response); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %v, 原始数据: %s", err, jsonStr)
	}

	return &response.Score, nil
}

// MergeFinalScore 合并最终分数
func (bs *BatchScorer) MergeFinalScore(score1, score2 int, score3 *int) float64 {
	if score3 == nil {
		// 无冲突，直接平均
		return float64(score1+score2) / 2.0
	}

	// 有冲突，选择最接近的两个分数的平均值
	scores := []int{score1, score2, *score3}
	sort.Ints(scores)

	// 计算相邻分数的差距
	diff1 := scores[1] - scores[0]
	diff2 := scores[2] - scores[1]

	var finalScore float64
	if diff1 <= diff2 {
		// 选择前两个分数的平均值
		finalScore = float64(scores[0]+scores[1]) / 2.0
		logrus.Infof("选择最接近的两个分数: %d, %d, 平均值: %.1f", scores[0], scores[1], finalScore)
	} else {
		// 选择后两个分数的平均值
		finalScore = float64(scores[1]+scores[2]) / 2.0
		logrus.Infof("选择最接近的两个分数: %d, %d, 平均值: %.1f", scores[1], scores[2], finalScore)
	}

	return finalScore
}
