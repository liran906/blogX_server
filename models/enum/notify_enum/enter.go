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

func (t Type) String() string {
	switch t {
	case ArticleLikeType:
		return "文章点赞"
	case ArticleCollectType:
		return "文章收藏"
	case ArticleCommentType:
		return "文章评论"
	case CommentLikeType:
		return "评论点赞"
	case CommentReplyType:
		return "评论回复"
	case SystemType:
		return "系统消息"
	}
	return "Unknown"
}
