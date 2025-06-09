// Path: ./models/user_model.go

package models

import (
	"blogX_server/models/enum"
	"time"
)

type UserModel struct {
	Model
	Username       string                  `gorm:"size:32; unique; not null" json:"username"`
	Email          string                  `gorm:"size:256; unique; not null" json:"email"`
	Password       string                  `gorm:"size:64" json:"-"`      // 密码可以null，比如通过 qq 注册
	PasswordUpdate int64                   `gorm:"default:null" json:"-"` // 密码更新时间 秒级时间戳
	Nickname       string                  `gorm:"size:32; not null" json:"nickname"`
	NicknameUpdate int64                   `gorm:"default:null" json:"-"` // 昵称更新时间 秒级时间戳
	AvatarURL      string                  `gorm:"size:256" json:"avatarURL"`
	Bio            string                  `gorm:"size:256" json:"bio"`
	OpenID         string                  `gorm:"size:64" json:"openid"`
	Gender         int8                    `json:"gender"`
	Phone          string                  `gorm:"size:16" json:"phone"`
	Country        string                  `gorm:"size:16" json:"country"`
	Province       string                  `gorm:"size:16" json:"province"`
	City           string                  `gorm:"size:16" json:"city"`
	Status         int8                    `json:"status"`
	LastLoginTime  time.Time               `json:"lastLoginTime"`
	LastLoginIP    string                  `gorm:"size:32" json:"lastLoginIP"`
	RegisterSource enum.RegisterSourceType `gorm:"not null" json:"registerSource"`
	DateOfBirth    time.Time               `gorm:"default:null" json:"dateOfBirth"`
	Role           enum.RoleType           `gorm:"not null" json:"role"` // 角色 1管理员 2普通用户 3访客

	// FK
	UserConfigModel      *UserConfigModel      `gorm:"foreignKey:UserID;references:ID" json:"-"` // 注意是指针，否则会报错：嵌套循环
	UserMessageConfModel *UserMessageConfModel `gorm:"foreignKey:UserID;references:ID" json:"-"`
	Articles             []ArticleModel        `gorm:"foreignKey:UserID" json:"-"`

	// M2M
	Images []ImageModel `gorm:"many2many:user_upload_images;joinForeignKey:UserID;JoinReferences:ImageID" json:"images"`
}

type UserConfigModel struct {
	UserID             uint       `gorm:"primaryKey" json:"userID"`
	UpdatedAt          *time.Time `json:"updatedAt"`                                        // 上次修改时间，可能为空，所以是指针
	Tags               []string   `gorm:"type:longtext; serializer:json" json:"tags"`       // 兴趣标签
	ThemeID            uint8      `gorm:"not null; default:1" json:"themeID"`               // 主页样式 id
	DisplayCollections bool       `gorm:"not null; default:true" json:"displayCollections"` // 公开我的收藏
	DisplayFans        bool       `gorm:"not null; default:true" json:"displayFans"`        // 公开我的粉丝
	DisplayFollowing   bool       `gorm:"not null; default:true" json:"displayFollowing"`   // 公开我的关注

	// FK
	UserModel UserModel `gorm:"foreignKey:UserID;references:ID" json:"userModel"` // 外键关联到 User, ref 如果不写会自动关联到 ID
}

func (u *UserModel) SiteAge() uint {
	return uint(time.Now().Sub(u.CreatedAt).Hours() / 24 / 365)
}
