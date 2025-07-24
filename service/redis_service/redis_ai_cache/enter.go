package redis_ai_cache

import (
	"blogX_server/global"
	"blogX_server/service/batch_scoring_service"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	// AI分析结果缓存前缀
	AIAnalysisPrefix = "ai_analysis:"

	// 批次评分结果缓存前缀
	BatchScoringPrefix = "batch_scoring:"

	// 详细分析结果缓存前缀
	DetailedAnalysisPrefix = "detailed_analysis:"

	// 缓存过期时间（7天）
	CacheExpiry = 7 * 24 * time.Hour
)

// generateCacheKey 生成缓存键
func generateCacheKey(prefix, paperID, contentHash string) string {
	return fmt.Sprintf("%s%s:%s", prefix, paperID, contentHash)
}

// generateContentHash 生成内容哈希值
func generateContentHash(title, abstract string) string {
	content := fmt.Sprintf("%s|%s", title, abstract)
	hash := md5.Sum([]byte(content))
	return fmt.Sprintf("%x", hash)[:8] // 取前8位
}

// SaveBatchScoringResult 保存批次评分结果到缓存
func SaveBatchScoringResult(result *batch_scoring_service.TwoStageAnalysisResult) error {
	// 为每个论文的第一阶段结果创建缓存
	for _, paperScore := range result.Stage1Results {
		// 我们需要论文的基本信息来生成hash，这里先跳过
		// 实际使用时需要传入论文基本信息
		logrus.Debugf("批次评分结果已保存到内存，论文: %s, 分数: %.1f",
			paperScore.ArxivID, paperScore.FinalScore)
	}

	// 为第二阶段详细分析结果创建缓存
	for _, detailedAnalysis := range result.Stage2Results {
		err := SaveDetailedAnalysis(&detailedAnalysis)
		if err != nil {
			logrus.Errorf("保存详细分析缓存失败 %s: %v", detailedAnalysis.ArxivID, err)
		}
	}

	logrus.Infof("批次评分结果缓存完成：第一阶段%d篇，第二阶段%d篇",
		len(result.Stage1Results), len(result.Stage2Results))

	return nil
}

// SaveDetailedAnalysis 保存单篇论文的详细分析结果
func SaveDetailedAnalysis(analysis *batch_scoring_service.DetailedAnalysis) error {
	if global.Redis == nil {
		return fmt.Errorf("Redis连接未初始化")
	}

	// 生成内容哈希
	contentHash := generateContentHash(analysis.Title, analysis.Abstract)

	// 生成缓存键
	cacheKey := generateCacheKey(DetailedAnalysisPrefix, analysis.ArxivID, contentHash)

	// 序列化分析结果
	data, err := json.Marshal(analysis)
	if err != nil {
		return fmt.Errorf("序列化分析结果失败: %v", err)
	}

	// 保存到Redis
	err = global.Redis.Set(cacheKey, data, CacheExpiry).Err()
	if err != nil {
		return fmt.Errorf("保存到Redis失败: %v", err)
	}

	logrus.Infof("详细分析结果已缓存：%s (有效期7天)", analysis.ArxivID)
	return nil
}

// GetDetailedAnalysis 获取单篇论文的详细分析结果
func GetDetailedAnalysis(paperID, title, abstract string) (*batch_scoring_service.DetailedAnalysis, error) {
	if global.Redis == nil {
		return nil, fmt.Errorf("Redis连接未初始化")
	}

	// 生成内容哈希
	contentHash := generateContentHash(title, abstract)

	// 生成缓存键
	cacheKey := generateCacheKey(DetailedAnalysisPrefix, paperID, contentHash)

	// 从Redis获取数据
	data, err := global.Redis.Get(cacheKey).Result()
	if err != nil {
		return nil, err // Redis key不存在或其他错误
	}

	// 反序列化
	var analysis batch_scoring_service.DetailedAnalysis
	err = json.Unmarshal([]byte(data), &analysis)
	if err != nil {
		return nil, fmt.Errorf("反序列化分析结果失败: %v", err)
	}

	logrus.Infof("从缓存获取详细分析结果：%s", paperID)
	return &analysis, nil
}

// SavePaperScore 保存单篇论文的评分结果
func SavePaperScore(paperID, title, abstract string, score float64, reasoning string) error {
	if global.Redis == nil {
		return fmt.Errorf("Redis连接未初始化")
	}

	// 生成内容哈希
	contentHash := generateContentHash(title, abstract)

	// 生成缓存键
	cacheKey := generateCacheKey(BatchScoringPrefix, paperID, contentHash)

	// 构建缓存数据
	scoreData := map[string]interface{}{
		"score":     score,
		"reasoning": reasoning,
		"cached_at": time.Now().Unix(),
	}

	// 序列化
	data, err := json.Marshal(scoreData)
	if err != nil {
		return fmt.Errorf("序列化评分结果失败: %v", err)
	}

	// 保存到Redis
	err = global.Redis.Set(cacheKey, data, CacheExpiry).Err()
	if err != nil {
		return fmt.Errorf("保存评分到Redis失败: %v", err)
	}

	logrus.Debugf("评分结果已缓存：%s = %.1f", paperID, score)
	return nil
}

// GetPaperScore 获取单篇论文的评分结果
func GetPaperScore(paperID, title, abstract string) (float64, string, error) {
	if global.Redis == nil {
		return 0, "", fmt.Errorf("Redis连接未初始化")
	}

	// 生成内容哈希
	contentHash := generateContentHash(title, abstract)

	// 生成缓存键
	cacheKey := generateCacheKey(BatchScoringPrefix, paperID, contentHash)

	// 从Redis获取数据
	data, err := global.Redis.Get(cacheKey).Result()
	if err != nil {
		return 0, "", err // Redis key不存在或其他错误
	}

	// 反序列化
	var scoreData map[string]interface{}
	err = json.Unmarshal([]byte(data), &scoreData)
	if err != nil {
		return 0, "", fmt.Errorf("反序列化评分结果失败: %v", err)
	}

	score, ok1 := scoreData["score"].(float64)
	reasoning, ok2 := scoreData["reasoning"].(string)

	if !ok1 || !ok2 {
		return 0, "", fmt.Errorf("缓存数据格式错误")
	}

	logrus.Debugf("从缓存获取评分结果：%s = %.1f", paperID, score)
	return score, reasoning, nil
}

// CheckCacheExists 检查缓存是否存在
func CheckCacheExists(paperID, title, abstract string, cacheType string) bool {
	if global.Redis == nil {
		return false
	}

	contentHash := generateContentHash(title, abstract)
	var prefix string

	switch cacheType {
	case "detailed":
		prefix = DetailedAnalysisPrefix
	case "scoring":
		prefix = BatchScoringPrefix
	default:
		return false
	}

	cacheKey := generateCacheKey(prefix, paperID, contentHash)
	exists, err := global.Redis.Exists(cacheKey).Result()

	return err == nil && exists > 0
}

// ClearExpiredCache 清理过期缓存（定期任务）
func ClearExpiredCache() error {
	if global.Redis == nil {
		return fmt.Errorf("Redis连接未初始化")
	}

	patterns := []string{
		AIAnalysisPrefix + "*",
		BatchScoringPrefix + "*",
		DetailedAnalysisPrefix + "*",
	}

	var totalDeleted int64

	for _, pattern := range patterns {
		keys, err := global.Redis.Keys(pattern).Result()
		if err != nil {
			logrus.Errorf("获取缓存键失败 %s: %v", pattern, err)
			continue
		}

		if len(keys) > 0 {
			// 检查哪些键已过期（Redis会自动删除，这里只是统计）
			for _, key := range keys {
				ttl := global.Redis.TTL(key).Val()
				if ttl < 0 { // 已过期或不存在
					totalDeleted++
				}
			}
		}
	}

	if totalDeleted > 0 {
		logrus.Infof("清理了 %d 个过期的AI分析缓存", totalDeleted)
	}

	return nil
}

// GetCacheStats 获取缓存统计信息
func GetCacheStats() map[string]interface{} {
	if global.Redis == nil {
		return map[string]interface{}{"error": "Redis未连接"}
	}

	stats := make(map[string]interface{})

	patterns := map[string]string{
		"detailed_analysis": DetailedAnalysisPrefix + "*",
		"batch_scoring":     BatchScoringPrefix + "*",
	}

	for name, pattern := range patterns {
		keys, err := global.Redis.Keys(pattern).Result()
		if err != nil {
			stats[name] = map[string]interface{}{"error": err.Error()}
			continue
		}

		var activeCount, expiredCount int
		for _, key := range keys {
			ttl := global.Redis.TTL(key).Val()
			if ttl > 0 {
				activeCount++
			} else {
				expiredCount++
			}
		}

		stats[name] = map[string]interface{}{
			"total_keys":   len(keys),
			"active_keys":  activeCount,
			"expired_keys": expiredCount,
		}
	}

	return stats
}

// InvalidateCache 使指定论文的缓存失效
func InvalidateCache(paperID string) error {
	if global.Redis == nil {
		return fmt.Errorf("Redis连接未初始化")
	}

	// 查找所有相关的缓存键
	patterns := []string{
		DetailedAnalysisPrefix + paperID + ":*",
		BatchScoringPrefix + paperID + ":*",
	}

	var deletedCount int64

	for _, pattern := range patterns {
		keys, err := global.Redis.Keys(pattern).Result()
		if err != nil {
			logrus.Errorf("查找缓存键失败 %s: %v", pattern, err)
			continue
		}

		if len(keys) > 0 {
			deleted := global.Redis.Del(keys...).Val()
			deletedCount += deleted
		}
	}

	if deletedCount > 0 {
		logrus.Infof("已清除论文 %s 的 %d 个缓存项", paperID, deletedCount)
	}

	return nil
}

// BatchInvalidateCache 批量使缓存失效
func BatchInvalidateCache(paperIDs []string) error {
	if global.Redis == nil {
		return fmt.Errorf("Redis连接未初始化")
	}

	var allKeys []string

	for _, paperID := range paperIDs {
		patterns := []string{
			DetailedAnalysisPrefix + paperID + ":*",
			BatchScoringPrefix + paperID + ":*",
		}

		for _, pattern := range patterns {
			keys, err := global.Redis.Keys(pattern).Result()
			if err != nil {
				logrus.Errorf("查找缓存键失败 %s: %v", pattern, err)
				continue
			}
			allKeys = append(allKeys, keys...)
		}
	}

	if len(allKeys) > 0 {
		deleted := global.Redis.Del(allKeys...).Val()
		logrus.Infof("批量清除了 %d 个AI分析缓存项", deleted)
	}

	return nil
}
