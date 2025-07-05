// Path: ./models/user_focus_model.go

package models

type UserFocusModel struct {
	Model
	UserID      uint `gorm:"uniqueIndex:idx_uniq_focus_uid" json:"userID"`      // 用户id
	FocusUserID uint `gorm:"uniqueIndex:idx_uniq_focus_uid" json:"focusUserID"` // 关注的用户

	// FK
	UserModel      UserModel `gorm:"foreignKey:UserID" json:"-"`
	FocusUserModel UserModel `gorm:"foreignKey:FocusUserID" json:"-"`
}
