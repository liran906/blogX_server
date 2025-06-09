// Path: ./models/enum/notify_enum/enter.go

package notify_enum

type Type int8

const (
	ArticleLikeType    Type = 1
	ArticleCollectType Type = 2
	ArticleCommentType Type = 3
	CommentLikeType    Type = 4
	CommentReplyType   Type = 5
	SystemType         Type = 6

	ArticleUnlikeType    Type = 7
	ArticleUncollectType Type = 8
	CommentUnlikeType    Type = 9
)
