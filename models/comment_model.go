package models

type CommentModel struct {
	Model
	Content    string `gorm:"not null" json:"content"`
	UserID     uint   `gorm:"index;not null" json:"userID"`
	ArticleID  uint   `gorm:"index;not null" json:"articleID"`
	ParentID   *uint  `gorm:"index" json:"parentID"`                 // 父评论
	RootID     *uint  `gorm:"index" json:"rootID"`                   // 根评论 自己为根时为 nil
	Depth      int    `gorm:"not null" json:"depth"`                 // 评论深度
	ReplyCount int    `gorm:"not null; default:0" json:"replyCount"` // 所有子孙回复数
	LikeCount  int    `gorm:"not null; default:0" json:"likeCount"`  // 点赞数

	// FK
	UserModel      UserModel       `gorm:"foreignKey:UserID;references:ID" json:"-"`
	ArticleModel   ArticleModel    `gorm:"foreignKey:ArticleID;references:ID" json:"-"`
	ParentModel    *CommentModel   `gorm:"foreignKey:ParentID;references:ID" json:"-"`
	RootModel      *CommentModel   `gorm:"foreignKey:RootID;references:ID" json:"-"`
	ChildListModel []*CommentModel `gorm:"foreignKey:ParentID" json:"childList"`
}
