// Path: ./models/user_global_notification_model.go

package models

// UserGlobalNotificationModel (U-GNM) 这个是配合 GlobalNotificationModel (GNM) 使用的
//
// GNM 中的是所有的全局消息 U-GNM 中是和用户关联过的
// 出现在 U-GNM 中的消息都是已读的，U-GNM 删除为 ture 则是这个用户删除了本条全局消息
// 如果对于某用户，一个全局消息：
// 1-未读未删 则在 U-GNM 表中没有记录（只在 GNM 表中有记录）
// 2-已读未删 则在 U-GNM 表中有记录且 IsDeleted=false
// 3-未读已删 则在 U-GNM 表中有记录且 IsDeleted=true (其实不会出现这个状态，因为只要在 U-GNM 表中就代表已读)
// 4-已读已删 则在 U-GNM 表中有记录且 IsDeleted=true
type UserGlobalNotificationModel struct {
	UserID               uint `gorm:"primaryKey" json:"userID"`
	GlobalNotificationID uint `gorm:"primaryKey" json:"globalNotificationID"`
	IsDeleted            bool `gorm:"not null;default:false" json:"isDeleted"`

	// FK
	UserModel               UserModel               `gorm:"foreignKey:UserID;references:ID" json:"-"`
	GlobalNotificationModel GlobalNotificationModel `gorm:"foreignKey:GlobalNotificationID;references:ID;constraint:OnDelete:CASCADE;" json:"-"`
}
