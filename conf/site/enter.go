// Path: ./conf/site/enter.go

package site

// SiteInfo 网站设置
type SiteInfo struct {
	Title        string `yaml:"title" json:"title"`                   // 标题
	EnglishTitle string `yaml:"englishTitle" json:"englishTitle"`     // 英文标题
	Slogan       string `yaml:"slogan" json:"slogan"`                 // 标语
	LogoURL      string `yaml:"logoURL" json:"logoURL"`               // logo url
	EnableICP    bool   `yaml:"enableICP" json:"enableICP"`           // 是否开启 icp 备案
	ICP          string `yaml:"icp" json:"icp"`                       // 备案号
	Mode         int8   `yaml:"mode" json:"mode" binding:"oneof=1 2"` // 1 社区模式 2 博客模式
}

// Project 项目设置
type Project struct {
	Title   string `yaml:"title" json:"title"`     // 网站标签 title
	Icon    string `yaml:"icon" json:"icon"`       // 图标
	WebPath string `yaml:"webPath" json:"webPath"` // 前端地址
}

// Seo 爬虫？
type Seo struct {
	Keywords    string `yaml:"keywords" json:"keywords"`
	Description string `yaml:"description" json:"description"`
}

// About 关于我们
type About struct {
	SiteDate  string `yaml:"siteDate" json:"siteDate"`   // 年月日
	QQURL     string `yaml:"qqURL" json:"qqURL"`         // QQ二维码
	Version   string `yaml:"-" json:"version"`           // 版本，要写死在代码中，所以 yaml 为 -
	WechatURL string `yaml:"wechatURL" json:"wechatURL"` // 微信二维码
	Gitee     string `yaml:"gitee" json:"gitee"`         // 网址
	Bilibili  string `yaml:"bilibili" json:"bilibili"`   // 网址
	Github    string `yaml:"github" json:"github"`       // 网址
}

// Login 登录设置
type Login struct {
	QQLogin          bool `yaml:"qqLogin" json:"qqLogin"`                   // qq 登录
	UsernamePwdLogin bool `yaml:"usernamePwdLogin" json:"usernamePwdLogin"` // 用户名密码登录
	EmailRegister    bool `yaml:"emailRegister" json:"emailRegister"`       // 邮箱登录
	Captcha          bool `yaml:"captcha" json:"captcha"`                   // 图片验证码
}

// IndexRight 右边栏设置
type IndexRight struct {
	List []ComponentInfo `json:"list" yaml:"list"`
}

// ComponentInfo 右边栏单个组件设置
type ComponentInfo struct {
	Title  string `yaml:"title" json:"title"`
	Enable bool   `yaml:"enable" json:"enable"`
}

// Article 文章设置
type Article struct {
	AutoApprove  bool `yaml:"autoApprove" json:"autoApprove"`   // 免审核
	MaxPin       int  `yaml:"maxPin" json:"maxPin"`             // 用户最高置顶数
	CommentDepth int  `yaml:"commentDepth" json:"commentDepth"` // 评论的层级
}

type AutoGen struct {
	UserID uint `yaml:"userID" json:"userID"`
}
