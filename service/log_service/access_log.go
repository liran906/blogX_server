package log_service

import (
	"blogX_server/core"
	"blogX_server/utils/jwts"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type AccessLog struct {
	c            *gin.Context
	responseBody []byte
	mutex        sync.Mutex
}

func NewAccessLog(c *gin.Context) *AccessLog {
	return &AccessLog{
		c: c,
	}
}

func (a *AccessLog) SetResponse(data []byte) {
	a.responseBody = data
}

func (a *AccessLog) Save() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	// 收集访问信息
	ip := a.c.ClientIP()
	ua := a.c.Request.UserAgent()
	method := a.c.Request.Method
	url := a.c.Request.URL.String()
	statusCode := a.c.Writer.Status()
	timestamp := time.Now()

	// 获取IP归属地
	addr, _ := core.GetLocationFromIP(ip)

	// 获取用户ID（如果已登录）
	var userID uint = 0
	claim, err := jwts.ParseTokenFromRequest(a.c)
	if claim != nil && err == nil {
		userID = claim.UserID
	}

	// 解析设备信息（简单解析User Agent）
	deviceInfo := a.parseUserAgent(ua)

	// 构造日志记录
	logEntry := fmt.Sprintf("[%s] IP:%s(%s) User:%d Method:%s URL:%s Status:%d Device:%s UA:%s",
		timestamp.Format("2006-01-02 15:04:05"),
		ip,
		addr,
		userID,
		method,
		url,
		statusCode,
		deviceInfo,
		ua,
	)

	// 写入文件
	if err := a.writeToFile(logEntry, timestamp); err != nil {
		logrus.Errorf("Failed to write access log: %v", err)
	}
}

func (a *AccessLog) parseUserAgent(ua string) string {
	ua = strings.ToLower(ua)
	
	// 检测移动设备
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		if strings.Contains(ua, "android") {
			return "Mobile-Android"
		} else if strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") {
			return "Mobile-iOS"
		}
		return "Mobile-Other"
	}

	// 检测桌面浏览器
	if strings.Contains(ua, "chrome") {
		return "Desktop-Chrome"
	} else if strings.Contains(ua, "firefox") {
		return "Desktop-Firefox"
	} else if strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome") {
		return "Desktop-Safari"
	} else if strings.Contains(ua, "edge") {
		return "Desktop-Edge"
	}

	return "Unknown"
}

func (a *AccessLog) writeToFile(logEntry string, timestamp time.Time) error {
	// 确保logs目录存在
	logsDir := "./logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return err
	}

	// 创建按日期分类的目录
	dateStr := timestamp.Format("2006-01-02")
	dailyDir := filepath.Join(logsDir, dateStr)
	if err := os.MkdirAll(dailyDir, 0755); err != nil {
		return err
	}

	// 访问日志文件路径
	logFilePath := filepath.Join(dailyDir, "access.log")

	// 以追加模式打开文件
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// 写入日志记录
	_, err = file.WriteString(logEntry + "\n")
	return err
}