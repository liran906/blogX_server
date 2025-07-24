package batch_scoring_service

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)

// BatchAllocator 批次分配器
type BatchAllocator struct {
	config *BatchScoringConfig
	rand   *rand.Rand
}

// NewBatchAllocator 创建批次分配器
func NewBatchAllocator(config *BatchScoringConfig) *BatchAllocator {
	return &BatchAllocator{
		config: config,
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// AllocatePapersToBatches 使用真正随机算法分配论文到批次
func (ba *BatchAllocator) AllocatePapersToBatches(paperCount int) (*BatchAllocation, error) {
	if paperCount <= 0 {
		return nil, fmt.Errorf("论文数量必须大于0")
	}

	// 1. 计算基本参数
	batchSize := ba.config.BatchSize
	totalPositions := paperCount * 2                             // 每篇论文需要2个位置
	totalBatches := (totalPositions + batchSize - 1) / batchSize // 向上取整

	logrus.Infof("批次分配开始：%d篇论文，每batch %d篇，需要%d个batch",
		paperCount, batchSize, totalBatches)

	// 2. 初始化分配结果
	allocation := &BatchAllocation{
		Batches:        make([][]int, totalBatches),
		PaperToBatches: make(map[int][]int),
		TotalBatches:   totalBatches,
	}

	// 初始化batch切片
	for i := 0; i < totalBatches; i++ {
		allocation.Batches[i] = make([]int, 0, batchSize)
	}

	// 3. 为每篇论文随机分配两个不同的batch
	for paperID := 0; paperID < paperCount; paperID++ {
		// 获取两个不同的随机batch
		batch1, batch2, err := ba.selectTwoDifferentBatches(totalBatches, allocation, batchSize)
		if err != nil {
			return nil, fmt.Errorf("为论文 %d 分配batch失败: %v", paperID, err)
		}

		// 分配到batch
		allocation.Batches[batch1] = append(allocation.Batches[batch1], paperID)
		allocation.Batches[batch2] = append(allocation.Batches[batch2], paperID)

		// 记录映射关系
		allocation.PaperToBatches[paperID] = []int{batch1, batch2}

		logrus.Debugf("论文 %d 分配到 batch %d 和 batch %d", paperID, batch1, batch2)
	}

	// 4. 验证分配结果
	if err := ba.validateAllocation(allocation, paperCount); err != nil {
		return nil, fmt.Errorf("分配验证失败: %v", err)
	}

	// 5. 输出分配统计
	ba.logAllocationStats(allocation)

	return allocation, nil
}

// selectTwoDifferentBatches 为一篇论文选择两个不同的batch，考虑负载均衡
func (ba *BatchAllocator) selectTwoDifferentBatches(totalBatches int, allocation *BatchAllocation, batchSize int) (int, int, error) {
	maxAttempts := 100 // 最大尝试次数，避免无限循环

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// 随机选择第一个batch
		batch1 := ba.rand.Intn(totalBatches)

		// 检查第一个batch是否还有空间
		if len(allocation.Batches[batch1]) >= batchSize {
			continue // 这个batch已满，尝试下一个
		}

		// 随机选择第二个batch
		batch2 := ba.rand.Intn(totalBatches)

		// 确保两个batch不同，且第二个batch还有空间
		if batch2 != batch1 && len(allocation.Batches[batch2]) < batchSize {
			return batch1, batch2, nil
		}
	}

	// 如果随机选择失败，使用确定性方法
	return ba.selectTwoBatchesDeterministic(totalBatches, allocation, batchSize)
}

// selectTwoBatchesDeterministic 确定性地选择两个可用的batch（备用方法）
func (ba *BatchAllocator) selectTwoBatchesDeterministic(totalBatches int, allocation *BatchAllocation, batchSize int) (int, int, error) {
	var availableBatches []int

	// 找到所有还有空间的batch
	for i := 0; i < totalBatches; i++ {
		if len(allocation.Batches[i]) < batchSize {
			availableBatches = append(availableBatches, i)
		}
	}

	if len(availableBatches) < 2 {
		return 0, 0, fmt.Errorf("可用batch数量不足，需要2个，实际有 %d 个", len(availableBatches))
	}

	// 随机选择两个可用的batch
	ba.shuffleSlice(availableBatches)
	return availableBatches[0], availableBatches[1], nil
}

// shuffleSlice 洗牌切片（Fisher-Yates算法）
func (ba *BatchAllocator) shuffleSlice(slice []int) {
	for i := len(slice) - 1; i > 0; i-- {
		j := ba.rand.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}

// validateAllocation 验证分配结果的正确性
func (ba *BatchAllocator) validateAllocation(allocation *BatchAllocation, paperCount int) error {
	// 1. 检查每篇论文是否恰好分配到2个batch
	for paperID := 0; paperID < paperCount; paperID++ {
		batches, exists := allocation.PaperToBatches[paperID]
		if !exists {
			return fmt.Errorf("论文 %d 没有分配到任何batch", paperID)
		}
		if len(batches) != 2 {
			return fmt.Errorf("论文 %d 分配到 %d 个batch，应该是2个", paperID, len(batches))
		}
		if batches[0] == batches[1] {
			return fmt.Errorf("论文 %d 被分配到同一个batch %d", paperID, batches[0])
		}
	}

	// 2. 检查batch大小是否合理
	for batchID, batch := range allocation.Batches {
		if len(batch) == 0 {
			return fmt.Errorf("batch %d 为空", batchID)
		}
		if len(batch) > ba.config.BatchSize {
			return fmt.Errorf("batch %d 大小 %d 超过限制 %d",
				batchID, len(batch), ba.config.BatchSize)
		}
	}

	// 3. 检查总位置数
	totalPositions := 0
	for _, batch := range allocation.Batches {
		totalPositions += len(batch)
	}
	expectedPositions := paperCount * 2
	if totalPositions != expectedPositions {
		return fmt.Errorf("总位置数 %d 不等于预期 %d", totalPositions, expectedPositions)
	}

	return nil
}

// logAllocationStats 输出分配统计信息
func (ba *BatchAllocator) logAllocationStats(allocation *BatchAllocation) {
	batchSizes := make([]int, len(allocation.Batches))
	minSize, maxSize := allocation.Batches[0], allocation.Batches[0]

	for i, batch := range allocation.Batches {
		batchSizes[i] = len(batch)
		if len(batch) < len(minSize) {
			minSize = batch
		}
		if len(batch) > len(maxSize) {
			maxSize = batch
		}
	}

	logrus.Infof("批次分配完成统计：")
	logrus.Infof("- 总批次数：%d", allocation.TotalBatches)
	logrus.Infof("- 最小批次大小：%d", len(minSize))
	logrus.Infof("- 最大批次大小：%d", len(maxSize))
	logrus.Infof("- 批次大小分布：%v", batchSizes)

	// 检查是否有同一篇论文在同一batch出现两次的情况
	duplicates := 0
	for batchID, batch := range allocation.Batches {
		seen := make(map[int]bool)
		for _, paperID := range batch {
			if seen[paperID] {
				logrus.Warnf("⚠️ batch %d 中论文 %d 出现重复", batchID, paperID)
				duplicates++
			}
			seen[paperID] = true
		}
	}

	if duplicates > 0 {
		logrus.Warnf("发现 %d 个重复分配", duplicates)
	} else {
		logrus.Infof("✅ 分配验证通过，无重复分配")
	}
}
