// Path: ./models/enum/message_enum/enter.go

package message_enum

type Type int8

const (
	ArticleLikeType      Type = 1
	ArticleUnlikeType    Type = 2
	ArticleCollectType   Type = 3
	ArticleUncollectType Type = 4
	ArticleCommentType   Type = 5
	CommentLikeType      Type = 6
	CommentUnlikeType    Type = 7
	CommentReplayType    Type = 8
	SystemType           Type = 9
)
