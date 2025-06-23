// Path: ./service/focus_service/enter.go

package focus_service

import (
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum/relationship_enum"
)

// CalcUserRelationship 计算好友关系
func CalcUserRelationship(A, B uint) (t relationship_enum.Relation) {
	//   2  用户2对用户1是什么关系
	if A == B {
		return relationship_enum.RelationSelf
	}
	var userFocusList []models.UserFocusModel
	global.DB.Find(&userFocusList,
		"(user_id = ? and focus_user_id = ? ) or (focus_user_id = ? and user_id = ? )",
		A, B, A, B)
	if len(userFocusList) == 2 {
		return relationship_enum.RelationFriend
	}
	if len(userFocusList) == 0 {
		return relationship_enum.RelationStranger
	}
	focus := userFocusList[0]
	if focus.FocusUserID == A {
		return relationship_enum.RelationFans
	}
	return relationship_enum.RelationFocus
}

// CalcUserPatchRelationship 批量计算好友关系
func CalcUserPatchRelationship(self uint, others []uint) (relationMap map[uint]relationship_enum.Relation) {
	relationMap = make(map[uint]relationship_enum.Relation)

	var relatedRelations []models.UserFocusModel
	global.DB.Find(&relatedRelations,
		"(user_id = ? and focus_user_id in ? ) or (focus_user_id = ? and user_id in ? )",
		self, others, self, others)

	for _, other := range others {
		if self == other {
			relationMap[other] = relationship_enum.RelationSelf
			continue
		}
		relationMap[other] = relationship_enum.RelationStranger
	}

	for _, relation := range relatedRelations {
		// A 关注了对方
		if relation.UserID == self {
			if relationMap[relation.FocusUserID] == relationship_enum.RelationFans {
				relationMap[relation.FocusUserID] = relationship_enum.RelationFriend
			} else {
				relationMap[relation.FocusUserID] = relationship_enum.RelationFocus
			}
		}
		// 对方关注了 A
		if relation.FocusUserID == self {
			if relationMap[relation.UserID] == relationship_enum.RelationFocus {
				relationMap[relation.UserID] = relationship_enum.RelationFriend
			} else {
				relationMap[relation.UserID] = relationship_enum.RelationFans
			}
		}
	}
	return
}
