package batch_scoring_service

import (
	"blogX_server/service/ai_service"
	"blogX_server/service/article_auto_generate/crawler_service"
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
	ArxivID    string // ArXiv ID
	Innovation int    // 创新性分数 (0-40)
	Technical  int    // 技术深度分数 (0-30)
	Practical  int    // 实用性分数 (0-30)
	Total      int    // 总分 (0-100)
}

// AI返回的JSON结构
type batchScoringAIResponse struct {
	Papers []struct {
		PaperID    string `json:"paper_id"`
		Innovation int    `json:"innovation"`
		Technical  int    `json:"technical"`
		Practical  int    `json:"practical"`
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

	//logrus.Debugf("BatchID=%d AI原始响应: %s", request.BatchID, rawResponse)

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

		total := aiResult.Innovation + aiResult.Technical + aiResult.Practical
		results = append(results, PaperBatchScore{
			ArxivID:    arxivID,
			Innovation: aiResult.Innovation,
			Technical:  aiResult.Technical,
			Practical:  aiResult.Practical,
			Total:      total,
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
		if result.Total < 0 || result.Total > 100 {
			return fmt.Errorf("论文 %s 的总分 %d 超出范围[0,100]", result.ArxivID, result.Total)
		}
		if result.Innovation < 0 || result.Innovation > 40 {
			return fmt.Errorf("论文 %s 的创新性分数 %d 超出范围[0,40]", result.ArxivID, result.Innovation)
		}
		if result.Technical < 0 || result.Technical > 30 {
			return fmt.Errorf("论文 %s 的技术深度分数 %d 超出范围[0,30]", result.ArxivID, result.Technical)
		}
		if result.Practical < 0 || result.Practical > 30 {
			return fmt.Errorf("论文 %s 的实用性分数 %d 超出范围[0,30]", result.ArxivID, result.Practical)
		}
	}

	// 检查分数区分度
	scores := make([]int, len(results))
	for i, result := range results {
		scores[i] = result.Total
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

// DetectScoreConflict 检测两个详细分数是否存在冲突
func (bs *BatchScorer) DetectScoreConflict(score1, score2 *DetailedScore) bool {
	if score1 == nil || score2 == nil {
		return false
	}

	// 检查总分差异
	totalDiff := int(math.Abs(float64(score1.Total - score2.Total)))
	if totalDiff > bs.config.ScoreDiffThreshold {
		return true
	}

	// 检查各项分数差异（相对阈值）
	innovationDiff := int(math.Abs(float64(score1.Innovation - score2.Innovation)))
	technicalDiff := int(math.Abs(float64(score1.Technical - score2.Technical)))
	practicalDiff := int(math.Abs(float64(score1.Practical - score2.Practical)))

	// 如果任何一项差异超过该项满分的50%，认为有冲突
	return innovationDiff > 20 || technicalDiff > 15 || practicalDiff > 15
}

// ScoreThirdRoundBatch 对需要第三次评分的论文进行批次评分
func (bs *BatchScorer) ScoreThirdRoundBatch(papers []crawler_service.ArxivPaper) (map[string]*DetailedScore, error) {
	if len(papers) == 0 {
		return make(map[string]*DetailedScore), nil
	}

	logrus.Infof("开始第三次批次评分：%d篇论文", len(papers))

	// 构建批次评分请求
	request := BatchScoringRequest{
		BatchID: 9999, // 特殊的第三次评分BatchID
		Papers:  papers,
		Attempt: 1,
	}

	// 执行批次评分
	response, err := bs.ScoreBatch(request)
	if err != nil {
		return nil, fmt.Errorf("第三次批次评分失败: %v", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("第三次批次评分不成功: %s", response.Error)
	}

	// 转换结果格式
	result := make(map[string]*DetailedScore)
	for _, batchScore := range response.Results {
		result[batchScore.ArxivID] = &DetailedScore{
			Innovation: batchScore.Innovation,
			Technical:  batchScore.Technical,
			Practical:  batchScore.Practical,
			Total:      batchScore.Total,
		}
	}

	logrus.Infof("第三次批次评分完成：成功评分%d篇论文", len(result))
	return result, nil
}

// MergeFinalScore 合并最终分数
func (bs *BatchScorer) MergeFinalScore(score1, score2 *DetailedScore, score3 *DetailedScore) float64 {
	if score1 == nil || score2 == nil {
		return 0.0
	}

	if score3 == nil {
		// 无冲突，直接平均
		return float64(score1.Total+score2.Total) / 2.0
	}

	// 有冲突，选择最接近的两个分数的平均值
	// 计算两两之间的总分差异
	diff12 := math.Abs(float64(score1.Total - score2.Total))
	diff13 := math.Abs(float64(score1.Total - score3.Total))
	diff23 := math.Abs(float64(score2.Total - score3.Total))

	var finalScore float64
	if diff12 <= diff13 && diff12 <= diff23 {
		// score1和score2最接近
		finalScore = float64(score1.Total+score2.Total) / 2.0
		logrus.Infof("选择最接近的两个分数: %d, %d, 平均值: %.1f", score1.Total, score2.Total, finalScore)
	} else if diff13 <= diff23 {
		// score1和score3最接近
		finalScore = float64(score1.Total+score3.Total) / 2.0
		logrus.Infof("选择最接近的两个分数: %d, %d, 平均值: %.1f", score1.Total, score3.Total, finalScore)
	} else {
		// score2和score3最接近
		finalScore = float64(score2.Total+score3.Total) / 2.0
		logrus.Infof("选择最接近的两个分数: %d, %d, 平均值: %.1f", score2.Total, score3.Total, finalScore)
	}

	return finalScore
}
