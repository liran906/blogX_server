package common_utils

import "fmt"

// getScoreEmoji 根据评分返回表情符号
func GetScoreEmoji(score float64) string {
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

// calculateAverage 计算平均值
func CalculateAverage(scores []int) float64 {
	if len(scores) == 0 {
		return 0
	}

	sum := 0
	for _, score := range scores {
		sum += score
	}

	return float64(sum) / float64(len(scores))
}

// CalculateAverageFloat 计算浮点数平均值
func CalculateAverageFloat(scores []float64) float64 {
	if len(scores) == 0 {
		return 0
	}

	sum := 0.0
	for _, score := range scores {
		sum += score
	}

	return sum / float64(len(scores))
}

// findMax 找最大值
func FindMax(scores []int) int {
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

// FindMaxFloat 找浮点数最大值
func FindMaxFloat(scores []float64) float64 {
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
func FindMin(scores []int) int {
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

// FindMinFloat 找浮点数最小值
func FindMinFloat(scores []float64) float64 {
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

// TruncateString 截断字符串到指定长度
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// TruncateAuthor 截断作者名称，专门用于作者字段
func TruncateAuthor(authors string, maxLen int) string {
	if len(authors) <= maxLen {
		return authors
	}
	return authors[:maxLen] + "..."
}

// FormatScoreRange 格式化分数范围，用于统计
func FormatScoreRange(score float64) string {
	rangeStart := int(score/10) * 10
	rangeEnd := rangeStart + 9
	return fmt.Sprintf("%d-%d", rangeStart, rangeEnd)
}
