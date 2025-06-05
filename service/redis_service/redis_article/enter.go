// Path: ./service/redis_service/redis_article/enter.go

package redis_article

type articleCacheType string

const (
	articleReadCount    articleCacheType = "article_read_count_key"
	articleLikeCount    articleCacheType = "article_like_count_key"
	articleCollectCount articleCacheType = "article_collect_count_key"
	articleCommentCount articleCacheType = "article_comment_count_key"
)
