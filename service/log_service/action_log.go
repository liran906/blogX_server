// Path: ./blogX_server/service/log_service/action_log.go

package log_service

import (
	"blogX_server/common/res"
	"blogX_server/core"
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"blogX_server/utils/jwts"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	e "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"reflect"
	"strings"
)

type ActionLog struct {
	c                  *gin.Context
	level              enum.LogLevelType
	title              string
	requestBody        []byte
	responseBody       []byte
	responseHeader     http.Header
	log                *models.LogModel // 备份存储，检测是否已存
	showRequestHeader  bool             // 是否显示响应头（不是所有视图都需要）
	showRequest        bool             // 是否显示请求体（不是所有视图都需要）
	showResponseHeader bool             // 是否显示响应头（不是所有视图都需要）
	showResponse       bool             // 是否显示响应体（不是所有视图都需要）
	itemList           []string         // 存放请求的content，最后和请求和响应body，合并为 content 入库
	isMiddlewareSave   bool             // 判断是否是在中间件中保存
}

func (l *ActionLog) ShowAll() {
	l.showRequestHeader = true
	l.showRequest = true
	l.showResponseHeader = true
	l.showResponse = true
}

func (l *ActionLog) ShowRequestHeader() {
	l.showRequestHeader = true
}

func (l *ActionLog) ShowRequest() {
	l.showRequest = true
}

func (l *ActionLog) ShowResponseHeader() {
	l.showResponseHeader = true
}

func (l *ActionLog) ShowResponse() {
	l.showResponse = true
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

	l.itemList = append(l.itemList, fmt.Sprintf("<div class=\"log_item %s\"><div class=\"log_item_label\">%s</div><div class=\"log_item_content\">%s</div></div>",
		logLevelType.ToString(), label, v))
}

func (l *ActionLog) SetItem(label string, value any) {
	l.setItem(label, value, enum.LogInfoLevel)
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

func (l *ActionLog) SetImage(src string) {
	l.itemList = append(l.itemList, fmt.Sprintf("<div class=\"log_image\"><img src=\"%s\" alt=\"\"></div>",
		src))
}

func (l *ActionLog) SetLink(label string, href string) {
	l.itemList = append(l.itemList, fmt.Sprintf("<div class=\"log_item link\"><div class=\"log_item_label\">%s</div><div class=\"log_item_content\"><a href=\"%s\" target=\"_blank\">%s</a></div></div> ",
		label, href, href))
}

func (l *ActionLog) SetError(label string, err error) {
	msg := e.WithStack(err)
	logrus.Errorf("%s: %s", label, err.Error())
	l.itemList = append(l.itemList, fmt.Sprintf("<div class=\"log_error\"><div class=\"line\"><div class=\"label\">%s</div><div class=\"value\">%s</div><div class=\"type\">%T</div></div><div class=\"stack\">%+v</div></div>",
		label, err, err, msg))
}

func (l *ActionLog) SetRequest(c *gin.Context) {
	// 读取 Body 并且回填到 Body（应对其阅后即焚的特性）
	byteData, err := c.GetRawData()
	if err != nil {
		logrus.Errorf(err.Error())
		return
	}
	// 写入到 Log 的RequestBody字段
	l.requestBody = byteData

	// Body 是阅后即焚（像 Python 中的迭代器） 所以要重新把内容回填进去
	c.Request.Body = io.NopCloser(bytes.NewReader(byteData))
}

func (l *ActionLog) SetResponseHeader(header http.Header) {
	l.responseHeader = header
}

func (l *ActionLog) SetResponse(data []byte) {
	l.responseBody = data
}

/*
Save 目前的问题：
	如果是在view 中去 save，那么是拿不到响应的，
	此时就算在响应中间件中再次去 save，响应数据也是没法写入的
所以解决为了解决这个问题：
	方法 1：只在响应中间件中调用 save。（不在 view 中调用）
	方法 2：在 view 中调用 save，需要返回日志的 id，其他接口可以根据 id 针对性修改这个 log 对象
*/

func (l *ActionLog) MiddlewareSave() {
	// 每一个中间件都会生成一个 action log 对象，但不一定每个 view 都需要一个 action log
	// 所以这里判断是否有 savedLog 字段（在 GetActionLog 方法中会设为 true）
	// 所以如果这里不为 true，则没有必要继续走下去了，直接 return
	_savedLog, _ := l.c.Get("savedLog") // 读取
	savedLog, _ := _savedLog.(bool)     // 断言
	if !savedLog {
		return
	}

	if l.log == nil {
		// 创建
		l.isMiddlewareSave = true
		l.Save()
		return
	}

	// 在 view 中 save 过，属于更新，要增加响应信息
	// 设置响应头
	if l.showResponseHeader {
		byteData, _ := json.Marshal(l.responseHeader)
		l.itemList = append(l.itemList, fmt.Sprintf("<div class=\"log_response_header\"><pre class=\"log_json_body\">%s</pre></div>",
			string(byteData)))
	}

	// 设置响应
	if l.showResponse {
		l.itemList = append(l.itemList, fmt.Sprintf("<div class=\"log_response\"><pre class=\"log_json_body\">%s</pre></div>",
			string(l.responseBody)))
	}

	// 然后走 save
	l.Save()
}

func (l *ActionLog) Save() uint {
	if l.log != nil {
		// Log 不为空，证明之前已经存过 Log 了，本次更新 content
		newContent := strings.Join(l.itemList, "\n")
		content := l.log.Content + "\n" + newContent
		global.DB.Model(l.log).Update("content", content)

		// 清空 itemList 以便下次保存增加新内容
		l.itemList = []string{}
		return l.log.ID
	}

	// 设置变量存储请求和响应的 相关 head 信息和 body
	var requestItemList []string

	// 设置请求头
	if l.showRequestHeader {
		byteData, _ := json.Marshal(l.c.Request.Header)
		requestItemList = append(requestItemList, fmt.Sprintf("<div class=\"log_request_header\"><pre class=\"log_json_body\">%s</pre></div>",
			string(byteData)))
	}

	// 设置请求
	if l.showRequest {
		requestItemList = append(requestItemList, fmt.Sprintf("<div class=\"log_request\"><div class=\"log_request_head\"><span class=\"log_request_method %s\">%s</span><span class=\"log_request_path\">%s</span></div><div class=\"log_request_body\"><pre class=\"log_json_body\">%s</pre></div></div>",
			strings.ToLower(l.c.Request.Method), l.c.Request.Method, l.c.Request.URL.String(), string(l.requestBody)))
	}

	// 合并中间的 content
	l.itemList = append(requestItemList, l.itemList...)

	// 在响应中间件中保存才有可能拿到响应信息
	if l.isMiddlewareSave {
		// 设置响应头
		if l.showResponseHeader {
			byteData, _ := json.Marshal(l.responseHeader)
			l.itemList = append(l.itemList, fmt.Sprintf("<div class=\"log_response_header\"><pre class=\"log_json_body\">%s</pre></div>", string(byteData)))
		}

		// 设置响应
		if l.showResponse {
			l.itemList = append(l.itemList, fmt.Sprintf("<div class=\"log_response\"><pre class=\"log_json_body\">%s</pre></div>", string(l.responseBody)))
		}
	}

	ip := l.c.ClientIP()
	ua := l.c.Request.UserAgent()
	addr, _ := core.GetLocationFromIP(ip)

	// 从 token 中读取 uid
	var userID uint
	claim, err := jwts.ParseTokenFromGin(l.c)
	if claim != nil && err == nil {
		userID = claim.UserID
	} else {
		// 这里按照教程没有，但我觉得应该报错+终止函数
		logrus.Errorf("failed to parse token: %v\n", err)
		res.FailWithError(err, l.c)
		return 0
	}

	log := models.LogModel{
		LogType:    enum.ActionLogType,
		Title:      l.title,
		Content:    strings.Join(l.itemList, "\n"), // 请求+content+响应，换行分割
		Level:      l.level,
		UserID:     userID,
		IP:         ip,
		IPLocation: addr,
		IsRead:     false,
		UA:         ua,
	}

	// 入库
	err = global.DB.Create(&log).Error
	if err != nil {
		logrus.Errorf("failed to create Log: %s\n", err)
	}

	// 写入 Log 字段（作为提醒已存过）
	l.log = &log

	// 清空 itemList 以便下次保存增加新内容
	l.itemList = []string{}

	return log.ID
}

func NewActionLogByGin(c *gin.Context) *ActionLog {
	return &ActionLog{c: c}
}

// GetActionLog 拿取对应 gin.Context 的 ActionLog 对象（如有）
func GetActionLog(c *gin.Context) *ActionLog {
	_log, ok := c.Get("log")
	if !ok {
		return NewActionLogByGin(c)
	}
	log, ok := _log.(*ActionLog)
	if !ok {
		return NewActionLogByGin(c)
	}

	// 每一个中间件都会生成一个 action log 对象，但不一定每个 view 都需要一个 action log
	// 所以这里定义一个 savedLog 字段，在响应中间件中可以检测，若不为 true，则不用保存 action log
	// 也就是说，只有调用了这个 GetActionLog 方法，中间件的 save 才会保存 action log
	c.Set("savedLog", true)

	return log
}
