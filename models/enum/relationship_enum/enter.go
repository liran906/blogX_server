// Path: ./models/enum/relationship_enum/enter.go

package relationship_enum

type Relation int8

//陌生人——双方都没有关注
//已关注——关注了对方，但是对方没有关注你
//粉丝——对方关注了你
//好友——双方互关

const (
	RelationStranger Relation = 1
	RelationFocus    Relation = 2
	RelationFans     Relation = 3
	RelationFriend   Relation = 4
	RelationSelf     Relation = 5
)
