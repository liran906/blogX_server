package models

type UserLoginModel struct {
	Model
	UserID     uint   `gorm:"not null" json:"userID"`
	IP         string `gorm:"size:32; not null" json:"ip"`
	IPLocation string `gorm:"size:64; not null" json:"ipLocation"`
	UA         string `gorm:"size:256; not null" json:"ua"`

	// FK
	UserModel UserModel `gorm:"foreignKey:UserID;references:ID" json:"-"`
}
