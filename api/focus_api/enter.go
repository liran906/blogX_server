// Path: ./api/focus_api/enter.go

package focus_api

import (
	"blogX_server/models/enum/relationship_enum"
	"time"
)

type FocusApi struct{}

type UserListResponse struct {
	UserID       uint                       `json:"userID"`
	UserNickname string                     `json:"userNickname"`
	UserAvatar   string                     `json:"userAvatar"`
	UserAbstract string                     `json:"userAbstract"`
	Relationship relationship_enum.Relation `json:"relationship"`
	CreatedAt    time.Time                  `json:"createdAt"`
}
