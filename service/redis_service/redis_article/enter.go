// Path: ./service/redis_service/redis_article/enter.go

package redis_article

type articleCacheType string

const (
	ArticleReadCount    articleCacheType = "article_read_count_key"
	ArticleLikeCount    articleCacheType = "article_like_count_key"
	ArticleCollectCount articleCacheType = "article_collect_count_key"
	ArticleCommentCount articleCacheType = "article_comment_count_key"
)
