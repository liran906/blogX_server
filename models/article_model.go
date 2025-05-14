package models

type ArticleModel struct {
	Model
	Title          string   `gorm:"size:128; not null" json:"title"`
	Abstract       string   `gorm:"size:256" json:"abstract"`
	CoverURL       string   `gorm:"size:256" json:"coverURL"`
	Content        string   `gorm:"not null" json:"content"`
	CategoryID     uint     `gorm:"not null" json:"categoryID"`
	Tags           []string `gorm:"type:longtext; serializer:json" json:"tags"`
	UserID         uint     `gorm:"not null" json:"userID"`
	Status         int8     `gorm:"not null" json:"status"` // 草稿 审核中 已发布
	ReadCount      int      `gorm:"not null" json:"readCount"`
	LikeCount      int      `gorm:"not null" json:"likeCount"`
	CommentCount   int      `gorm:"not null" json:"commentCount"`
	OpenForComment bool     `gorm:"not null; default:true" json:"openForComment"`

	// FK
	UserModel UserModel `gorm:"foreignKey:UserID;references:ID" json:"-"`
}
