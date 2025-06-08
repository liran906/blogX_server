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
	err := global.Redis.Del(string(ArticleReadCount), string(ArticleLikeCount), string(ArticleCollectCount), string(ArticleCommentCount)).Err()
	if err != nil {
		logrus.Errorf("Failed to clear article redis cache: %v", err)
	}
}

// 增减更新数值

func UpdateArticleRead(articleID uint, delta int) {
	update(ArticleReadCount, articleID, delta)
}
func UpdateArticleLike(articleID uint, delta int) {
	update(ArticleLikeCount, articleID, delta)
}
func UpdateArticleCollect(articleID uint, delta int) {
	update(ArticleCollectCount, articleID, delta)
}
func UpdateArticleComment(articleID uint, delta int) {
	update(ArticleCommentCount, articleID, delta)
}

// 加一

func AddArticleRead(articleID uint) {
	update(ArticleReadCount, articleID, 1)
}
func AddArticleLike(articleID uint) {
	update(ArticleLikeCount, articleID, 1)
}
func AddArticleCollect(articleID uint) {
	update(ArticleCollectCount, articleID, 1)
}
func AddArticleComment(articleID uint) {
	update(ArticleCommentCount, articleID, 1)
}

// 减一

func SubArticleLike(articleID uint) {
	update(ArticleLikeCount, articleID, -1)
}
func SubArticleRead(articleID uint) {
	update(ArticleReadCount, articleID, -1)
}
func SubArticleCollect(articleID uint) {
	update(ArticleCollectCount, articleID, -1)
}
func SubArticleComment(articleID uint) {
	update(ArticleCommentCount, articleID, -1)
}

// 设值

func SetArticleRead(articleID uint, n int) {
	set(ArticleReadCount, articleID, n)
}
func SetArticleLike(articleID uint, n int) {
	set(ArticleLikeCount, articleID, n)
}
func SetArticleCollect(articleID uint, n int) {
	set(ArticleCollectCount, articleID, n)
}
func SetArticleComment(articleID uint, n int) {
	set(ArticleCommentCount, articleID, n)
}
