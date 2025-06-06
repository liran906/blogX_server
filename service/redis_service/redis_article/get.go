// Path: ./service/redis_service/redis_article/get.go

package redis_article

import (
	"blogX_server/global"
	"blogX_server/models"
	"strconv"
)

func get(t articleCacheType, articleID uint) int {
	num, _ := global.Redis.HGet(string(t), strconv.Itoa(int(articleID))).Int()
	return num
}

func GetArticleRead(articleID uint) int {
	return get(articleReadCount, articleID)
}
func GetArticleLike(articleID uint) int {
	return get(articleLikeCount, articleID)
}
func GetArticleCollect(articleID uint) int {
	return get(articleCollectCount, articleID)
}
func GetArticleComment(articleID uint) int {
	return get(articleCommentCount, articleID)
}

func getAllArticles(t articleCacheType) map[uint]int {
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

func GetAllReadCounts() map[uint]int {
	return getAllArticles(articleReadCount)
}
func GetAllLikeCounts() map[uint]int {
	return getAllArticles(articleLikeCount)
}
func GetAllCollectCounts() map[uint]int {
	return getAllArticles(articleCollectCount)
}
func GetAllCommentCounts() map[uint]int {
	return getAllArticles(articleCommentCount)
}

func GetAllTypes(articleID uint) map[string]int {
	mps := map[string]int{
		"read":    GetArticleRead(articleID),
		"like":    GetArticleLike(articleID),
		"collect": GetArticleCollect(articleID),
		"comment": GetArticleComment(articleID),
	}
	return mps
}
func GetAllTypesForArticle(article *models.ArticleModel) (ok bool) {
	if article == nil || article.ID == 0 {
		return false
	}
	mps := GetAllTypes(article.ID)
	article.ReadCount = mps["read"]
	article.LikeCount = mps["like"]
	article.CollectCount = mps["collect"]
	article.CommentCount = mps["comment"]
	return true
}
