// Path: ./service/redis_service/redis_comment/enter.go

package redis_comment

import (
	"blogX_server/global"
	"github.com/sirupsen/logrus"
	"strconv"
)

type commentCacheType string

const (
	commentReplyCount commentCacheType = "comment_reply_count_key"
	commentLikeCount  commentCacheType = "comment_like_count_key"
)

// 基本方法

func update(t commentCacheType, commentID uint, delta int) {
	global.Redis.HIncrBy(string(t), strconv.Itoa(int(commentID)), int64(delta))
}

func set(t commentCacheType, commentID uint, n int) {
	global.Redis.HSet(string(t), strconv.Itoa(int(commentID)), strconv.Itoa(n))
}

func get(t commentCacheType, commentID uint) int {
	num, _ := global.Redis.HGet(string(t), strconv.Itoa(int(commentID))).Int()
	return num
}

func Clear() {
	err := global.Redis.Del(string(commentReplyCount), string(commentLikeCount)).Err()
	if err != nil {
		logrus.Errorf("Failed to clear article redis cache: %v", err)
	}
}

// 增减更新数值

func UpdateCommentReplyCount(commentID uint, delta int) {
	update(commentReplyCount, commentID, delta)
}
func UpdateCommentLikeCount(commentID uint, delta int) {
	update(commentLikeCount, commentID, delta)
}

// 加一

func AddCommentReplyCount(commentID uint) {
	update(commentReplyCount, commentID, 1)
}
func AddCommentLikeCount(commentID uint) {
	update(commentLikeCount, commentID, 1)
}

// 减一

func SubCommentReplyCount(commentID uint) {
	update(commentReplyCount, commentID, -1)
}
func SubCommentLikeCount(commentID uint) {
	update(commentLikeCount, commentID, -1)
}

// 设值

func SetCommentReplyCount(commentID uint, n int) {
	set(commentReplyCount, commentID, n)
}
func SetCommentLikeCount(commentID uint, n int) {
	set(commentLikeCount, commentID, n)
}

// get

func GetCommentReplyCount(commentID uint) int {
	return get(commentReplyCount, commentID)
}
func GetCommentLikeCount(commentID uint) int {
	return get(commentLikeCount, commentID)
}
