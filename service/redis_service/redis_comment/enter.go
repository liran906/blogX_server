// Path: ./service/redis_service/redis_comment/enter.go

package redis_comment

import (
	"blogX_server/global"
	"github.com/sirupsen/logrus"
	"strconv"
)

type commentCacheType string

const (
	CommentReplyCount commentCacheType = "comment_reply_count_key"
	CommentLikeCount  commentCacheType = "comment_like_count_key"
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
	err := global.Redis.Del(string(CommentReplyCount), string(CommentLikeCount)).Err()
	if err != nil {
		logrus.Errorf("Failed to clear article redis cache: %v", err)
	}
}

// 增减更新数值

func UpdateCommentReplyCount(commentID uint, delta int) {
	update(CommentReplyCount, commentID, delta)
}
func UpdateCommentLikeCount(commentID uint, delta int) {
	update(CommentLikeCount, commentID, delta)
}

// 加一

func AddCommentReplyCount(commentID uint) {
	update(CommentReplyCount, commentID, 1)
}
func AddCommentLikeCount(commentID uint) {
	update(CommentLikeCount, commentID, 1)
}

// 减一

func SubCommentReplyCount(commentID uint) {
	update(CommentReplyCount, commentID, -1)
}
func SubCommentLikeCount(commentID uint) {
	update(CommentLikeCount, commentID, -1)
}

// 设值

func SetCommentReplyCount(commentID uint, n int) {
	set(CommentReplyCount, commentID, n)
}
func SetCommentLikeCount(commentID uint, n int) {
	set(CommentLikeCount, commentID, n)
}

// get

func GetCommentReplyCount(commentID uint) int {
	return get(CommentReplyCount, commentID)
}
func GetCommentLikeCount(commentID uint) int {
	return get(CommentLikeCount, commentID)
}

// get all

func getAllCommentCache(t commentCacheType) map[uint]int {
	res, err := global.Redis.HGetAll(string(t)).Result()
	if err != nil {
		return nil
	}
	mps := make(map[uint]int)
	for k, v := range res {
		key, err1 := strconv.Atoi(k)
		val, err2 := strconv.Atoi(v)
		if err1 != nil || err2 != nil {
			continue // skip this invalid entry
		}
		mps[uint(key)] = val
	}
	return mps
}

func GetAllReplyCounts() map[uint]int {
	return getAllCommentCache(CommentReplyCount)
}
func GetAllLikeCounts() map[uint]int {
	return getAllCommentCache(CommentLikeCount)
}
