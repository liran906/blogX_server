// Path: ./models/data_model.go

package models

import "time"

type DataModel struct {
	Date       time.Time `gorm:"primaryKey" json:"date"`
	FlowCount  int       `gorm:"not null" json:"flowCount"`
	ClickCount int       `gorm:"not null" json:"clickCount"`
}
