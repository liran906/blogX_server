package batch_scoring_service

import (
	"blogX_server/service/article_auto_generate/crawler_service"
)

// BatchScoringConfig 批量评分配置
type BatchScoringConfig struct {
	BatchSize           int // 每个batch的大小 (8-12)
	ScoreDiffThreshold  int // 触发第三次评分的分数差阈值 (20)
	MaxRetries          int // 单个batch最大重试次数 (3)
	TopN                int // 最终选择的top论文数量
	ThirdRoundBatchSize int // 第三次评分的batch大小 (6-12)
}

// DefaultBatchScoringConfig 默认配置
func DefaultBatchScoringConfig() *BatchScoringConfig {
	return &BatchScoringConfig{
		BatchSize:           10,
		ScoreDiffThreshold:  20,
		MaxRetries:          5,
		TopN:                20,
		ThirdRoundBatchSize: 8, // 第三次评分batch大小
	}
}

// DetailedScore 详细分项评分
type DetailedScore struct {
	Innovation int // 创新性 (0-40)
	Technical  int // 技术深度 (0-30)
	Practical  int // 实用性 (0-30)
	Total      int // 总分 (0-100)
}

// PaperScore 论文评分结果
type PaperScore struct {
	ArxivID    string         // 论文ID
	Score1     *DetailedScore // 第一次评分
	Score2     *DetailedScore // 第二次评分
	Score3     *DetailedScore // 第三次评分（可选）
	FinalScore float64        // 最终评分
	BatchIDs   []int          // 所属的batch ID列表
	Status     ScoreStatus    // 评分状态
}

// ScoreStatus 评分状态
type ScoreStatus int

const (
	StatusPending    ScoreStatus = iota // 待评分
	StatusPartial                       // 部分完成（只有一次评分）
	StatusCompleted                     // 完成（两次评分）
	StatusThirdRound                    // 需要第三次评分
	StatusFailed                        // 评分失败
)

// BatchAllocation 批次分配结果
type BatchAllocation struct {
	Batches        [][]int       // 每个batch包含的论文索引
	PaperToBatches map[int][]int // 论文到batch的映射
	TotalBatches   int           // 总batch数
}

// BatchScoringResult 单个batch的评分结果
type BatchScoringResult struct {
	BatchID  int            // batch ID
	Scores   map[string]int // 论文ID -> 评分
	Success  bool           // 是否成功
	Error    error          // 错误信息
	Attempts int            // 尝试次数
}

// TwoStageResult 两阶段分析的最终结果
type TwoStageResult struct {
	// 第一阶段：批量评分结果
	AllScores    []*PaperScore // 所有论文的评分
	FailedPapers []string      // 评分失败的论文ID

	// 第二阶段：详细分析结果
	TopPapers    []*PaperAnalysisResult // 高分论文的详细分析
	AnalysisTime string                 // 分析时间

	// 统计信息
	Stats *ScoringStats // 评分统计
}

// ScoringStats 评分统计信息
type ScoringStats struct {
	TotalPapers        int     // 总论文数
	SuccessfullyScored int     // 成功评分数
	FailedPapers       int     // 失败论文数
	NeedThirdRound     int     // 需要第三次评分数
	AvgScore           float64 // 平均分
	MaxScore           float64 // 最高分
	MinScore           float64 // 最低分
	TotalBatches       int     // 总batch数
	FailedBatches      int     // 失败batch数
}

// PaperAnalysisResult 详细分析结果（从autogen_service引用）
type PaperAnalysisResult struct {
	ArxivID          string   `json:"arxivId"`
	Title            string   `json:"title"`
	Authors          string   `json:"authors"`
	Score            int      `json:"score"`
	Justification    string   `json:"justification"`
	Tags             []string `json:"tags"`
	Abstract         string   `json:"abstract"`
	PublishedDate    string   `json:"publishedDate"`
	AnalyzedAt       string   `json:"analyzedAt"`
	OriginalAbstract string   `json:"originalAbstract"`
	PdfURL           string   `json:"pdfUrl"`
	HtmlURL          string   `json:"htmlUrl"`
}

// BatchScoringService 批量评分服务
type BatchScoringService struct {
	config *BatchScoringConfig
}

// NewBatchScoringService 创建批量评分服务
func NewBatchScoringService(config *BatchScoringConfig) *BatchScoringService {
	if config == nil {
		config = DefaultBatchScoringConfig()
	}

	return &BatchScoringService{
		config: config,
	}
}

// AnalyzePapersInTwoStages 两阶段分析主入口
func (s *BatchScoringService) AnalyzePapersInTwoStages(papers []crawler_service.ArxivPaper, category string) (*TwoStageResult, error) {
	// 实现将在two_stage_analyzer.go中
	return nil, nil
}
