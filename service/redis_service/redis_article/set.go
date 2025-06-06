// Path: ./service/redis_service/redis_article/enter.go

package redis_article

import (
	"blogX_server/global"
	"github.com/sirupsen/logrus"
	"strconv"
)

// 基本方法

func update(t articleCacheType, articleID uint, delta int) {
	global.Redis.HIncrBy(string(t), strconv.Itoa(int(articleID)), int64(delta))
}

func set(t articleCacheType, articleID uint, n int) {
	global.Redis.HSet(string(t), strconv.Itoa(int(articleID)), strconv.Itoa(n))
}

func Clear() {
	err := global.Redis.Del(string(articleReadCount), string(articleLikeCount), string(articleCollectCount), string(articleCommentCount)).Err()
	if err != nil {
		logrus.Errorf("Failed to clear article redis cache: %v", err)
	}
}

// 增减更新数值

func UpdateArticleRead(articleID uint, delta int) {
	update(articleReadCount, articleID, delta)
}
func UpdateArticleLike(articleID uint, delta int) {
	update(articleLikeCount, articleID, delta)
}
func UpdateArticleCollect(articleID uint, delta int) {
	update(articleCollectCount, articleID, delta)
}
func UpdateArticleComment(articleID uint, delta int) {
	update(articleCommentCount, articleID, delta)
}

// 加一

func AddArticleRead(articleID uint) {
	update(articleReadCount, articleID, 1)
}
func AddArticleLike(articleID uint) {
	update(articleLikeCount, articleID, 1)
}
func AddArticleCollect(articleID uint) {
	update(articleCollectCount, articleID, 1)
}
func AddArticleComment(articleID uint) {
	update(articleCommentCount, articleID, 1)
}

// 减一

func SubArticleLike(articleID uint) {
	update(articleLikeCount, articleID, -1)
}
func SubArticleRead(articleID uint) {
	update(articleReadCount, articleID, -1)
}
func SubArticleCollect(articleID uint) {
	update(articleCollectCount, articleID, -1)
}
func SubArticleComment(articleID uint) {
	update(articleCommentCount, articleID, -1)
}

// 设值

func SetArticleRead(articleID uint, n int) {
	set(articleReadCount, articleID, n)
}
func SetArticleLike(articleID uint, n int) {
	set(articleLikeCount, articleID, n)
}
func SetArticleCollect(articleID uint, n int) {
	set(articleCollectCount, articleID, n)
}
func SetArticleComment(articleID uint, n int) {
	set(articleCommentCount, articleID, n)
}
