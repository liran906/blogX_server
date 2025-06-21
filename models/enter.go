// Path: ./models/enter.go

package models

import (
	"time"
)

type Model struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type IDRequest struct {
	ID uint `uri:"id" json:"id" form:"id"`
}

type IDListRequest struct {
	IDList []uint `json:"idList"`
}

type OptionsRequest[T any] struct {
	Label string `json:"label"`
	Value T      `json:"value"`
}
