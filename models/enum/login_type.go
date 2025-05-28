package enum

type LoginType uint8

const (
	UsernamePasswordLoginType LoginType = 1
	EmailPasswordLoginType    LoginType = 2
	QQLoginType               LoginType = 3
	WechatLoginType           LoginType = 4
)

func (l LoginType) String() string {
	switch l {
	case UsernamePasswordLoginType:
		return "用户名密码登录"
	case EmailPasswordLoginType:
		return "邮箱密码登录"
	case QQLoginType:
		return "QQ登录"
	case WechatLoginType:
		return "微信登录"
	}
	return ""
}
