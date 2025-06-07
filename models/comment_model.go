package models

type CommentModel struct {
	Model
	Content   string `gorm:"not null" json:"content"`
	UserID    uint   `gorm:"not null" json:"userID"`
	ArticleID uint   `gorm:"not null" json:"articleID"`
	ParentID  *uint  `json:"parentID"` // 父评论
	RootID    *uint  `json:"rootID"`   // 根评论 自己为根时为 nil
	LikeCount int    `gorm:"not null" json:"likeCount"`

	// FK
	UserModel      UserModel       `gorm:"foreignKey:UserID;references:ID" json:"-"`
	ArticleModel   ArticleModel    `gorm:"foreignKey:ArticleID;references:ID" json:"-"`
	ParentModel    *CommentModel   `gorm:"foreignKey:ParentID;references:ID" json:"-"`
	RootModel      *CommentModel   `gorm:"foreignKey:RootID;references:ID" json:"-"`
	ChildListModel []*CommentModel `gorm:"foreignKey:ParentID" json:"childList"` // 这里不确定是否要 FK。按照 GPT 连字段都不用
}
