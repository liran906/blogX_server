// Path: ./models/enum/article_status.go

package enum

type ArticleStatus uint8

const (
	ArticleStatusDraft   ArticleStatus = 1
	ArticleStatusReview  ArticleStatus = 2
	ArticleStatusPublish ArticleStatus = 3
	AritcleStatusFail    ArticleStatus = 4
)
