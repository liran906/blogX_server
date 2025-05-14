package enum

type LoginType uint8

const (
	UsernamePasswordLoginType LoginType = 1
	EmailPasswordLoginType    LoginType = 2
	QQLoginType               LoginType = 3
	WechatLoginType           LoginType = 4
)
