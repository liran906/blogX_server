// Path: ./blogX_server/models/enter.go

package models

import (
	"time"
)

type Model struct {
	ID        uint `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type IDRequest struct {
	ID uint `uri:"id" json:"id" form:"id"`
}

type RemoveRequest struct {
	IDList []uint `json:"idList"`
}
