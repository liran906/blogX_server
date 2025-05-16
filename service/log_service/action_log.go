// Path: ./service/log_service/action_log.go

package log_service

import (
	"blogX_server/core"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io"
	"reflect"
	"strings"
)

type ActionLog struct {
	c            *gin.Context
	level        enum.LogLevelType
	title        string
	requestBody  []byte
	responseBody []byte
	log          *models.LogModel // 备份存储，检测是否已存
	showResponse bool             // 是否显示响应体（不是所有视图都需要）
	showRequest  bool             // 是否显示请求体（不是所有视图都需要）
	itemList     []string         // 存放请求的content，最后和请求和响应body，合并为 content 入库
}

func (l *ActionLog) SetShowResponse(show bool) {
	l.showResponse = show
}

func (l *ActionLog) SetShowRequest(show bool) {
	l.showRequest = show
}

func (l *ActionLog) SetLevel(level enum.LogLevelType) {
	l.level = level
}

func (l *ActionLog) SetTitle(title string) {
	l.title = title
}

func (l *ActionLog) setItem(label string, value any, logLevelType enum.LogLevelType) {
	// 用 reflect 判断 value 类型
	var v string
	t := reflect.TypeOf(value)
	switch t.Kind() {
	case reflect.Struct, reflect.Map, reflect.Slice: // 这三种类型需要转换为 json
		ByteData, _ := json.Marshal(value)
		v = string(ByteData) // tbd
	default:
		v = fmt.Sprintf("%v", value)
	}

	l.itemList = append(l.itemList, fmt.Sprintf("Log level: %s; Log label: %s; Value: %s",
		logLevelType.ToString(), label, v))
}

func (l *ActionLog) SetItem(label string, value any) {
	l.setItem(label, value, enum.LogDebugLevel)
}

func (l *ActionLog) SetItemDebug(label string, value any) {
	l.setItem(label, value, enum.LogDebugLevel)
}

func (l *ActionLog) SetItemInfo(label string, value any) {
	l.setItem(label, value, enum.LogInfoLevel)
}

func (l *ActionLog) SetItemWarn(label string, value any) {
	l.setItem(label, value, enum.LogWarnLevel)
}

func (l *ActionLog) SetItemError(label string, value any) {
	l.setItem(label, value, enum.LogErrorLevel)
}

func (l *ActionLog) SetItemFatal(label string, value any) {
	l.setItem(label, value, enum.LogFatalLevel)
}

func (l *ActionLog) SetItemPanic(label string, value any) {
	l.setItem(label, value, enum.LogPanicLevel)
}

func (l *ActionLog) SetRequest(c *gin.Context) {
	// 读取 Body 并且回填到 Body（应对其阅后即焚的特性）
	ByteData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logrus.Errorf(err.Error())
	}
	// 回填到 log 的RequestBody字段
	l.requestBody = ByteData

	// Body 是阅后即焚（像 Python 中的迭代器） 所以要重新把内容回填进去
	c.Request.Body = io.NopCloser(bytes.NewReader(ByteData))
}

func (l *ActionLog) SetResponse(data []byte) {
	l.responseBody = data
}

func (l *ActionLog) Save() {
	if l.log != nil {
		// log 不为空，证明之前已经存过 log 了
		global.DB.Model(l.log).Update("title", "test: 已更新") // tbd
		return
	}

	// 设置变量存储请求和响应的 相关 head 信息和 body
	var NewItemList []string

	// 设置请求
	if l.showRequest {
		NewItemList = append(NewItemList, fmt.Sprintf("request method: %s; request path: %s; request body: %s",
			l.c.Request.Method, l.c.Request.URL.String(), string(l.requestBody)))
	}

	// 合并中间的 content
	NewItemList = append(NewItemList, l.itemList...)

	// 设置响应
	if l.showResponse {
		NewItemList = append(NewItemList, fmt.Sprintf("response body: %s", string(l.responseBody)))
	}

	ip := l.c.ClientIP()
	ua := l.c.Request.UserAgent()
	addr, _ := core.GetAddress(ip)

	userID := uint(1) // tbd

	log := models.LogModel{
		LogType: enum.ActionLogType,
		Title:   l.title,
		Content: strings.Join(NewItemList, "\n"), // 请求+content+响应，换行分割
		Level:   l.level,
		UserID:  userID,
		IP:      ip,
		Address: addr,
		IsRead:  false,
		UA:      ua,
	}
	err := global.DB.Create(&log).Error
	if err != nil {
		logrus.Errorf("failed to create log: %s\n", err)
	}

	// 写入 log 字段（作为提醒已存过）
	l.log = &log
}

func NewActionLogByGin(c *gin.Context) *ActionLog {
	return &ActionLog{c: c}
}

func GetLog(c *gin.Context) *ActionLog {
	_log, ok := c.Get("log")
	if !ok {
		return NewActionLogByGin(c)
	}
	log, ok := _log.(*ActionLog)
	if !ok {
		return NewActionLogByGin(c)
	}
	return log
}
