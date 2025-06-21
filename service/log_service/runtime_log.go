// Path: ./service/log_service/runtime_log.go

package log_service

import (
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"encoding/json"
	"fmt"
	e "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"reflect"
	"strings"
	"time"
)

type RuntimeLog struct {
	title        string
	content      string
	level        enum.LogLevelType
	itemList     []string
	serviceName  string
	runtimeDelta RuntimeDelta
}

// RuntimeDelta 更新频率
type RuntimeDelta int8

const (
	RuntimeDeltaHour  RuntimeDelta = 1
	RuntimeDeltaDay   RuntimeDelta = 2
	RuntimeDeltaWeek  RuntimeDelta = 3
	RuntimeDeltaMonth RuntimeDelta = 4
)

func NewRuntimeLog(serviceName string, delta RuntimeDelta) *RuntimeLog {
	return &RuntimeLog{
		serviceName:  serviceName,
		runtimeDelta: delta,
		level:        enum.LogInfoLevel,
	}
}

func (l *RuntimeLog) GetSqlDelta() string {
	switch l.runtimeDelta {
	case RuntimeDeltaHour:
		return "INTERVAL 1 HOUR"
	case RuntimeDeltaDay:
		return "INTERVAL 1 DAY"
	case RuntimeDeltaWeek:
		return "INTERVAL 1 WEEK"
	case RuntimeDeltaMonth:
		return "INTERVAL 1 MONTH"
	default:
		return "INTERVAL 1 DAY"
	}
}

func (l *RuntimeLog) Save() {
	// 记录保存时间
	l.SetNowTime()

	//// 判断是更新还是创建
	//var log *models.LogModel
	//
	//global.DB.Find(&log, fmt.Sprintf("service_name = ? and log_type = %d and created_at >= date_sub(now(), %s)",
	//	enum.RuntimeLogType, l.GetSqlDelta()), l.serviceName)
	//
	//// 更新逻辑
	//if log.ID != 0 {
	//	// 更新 content
	//	newContent := strings.Join(l.itemList, "\n")
	//	content := log.Content + "\n" + newContent
	//	global.DB.Model(log).Update("content", content)
	//
	//	// 清空 itemList 以便下次保存增加新内容
	//	l.itemList = []string{}
	//	return
	//}

	// 创建逻辑
	log := &models.LogModel{
		Title:       l.title,
		Content:     strings.Join(l.itemList, "\n"),
		Level:       l.level,
		LogType:     enum.RuntimeLogType,
		ServiceName: l.serviceName,
	}
	err := global.DB.Create(log).Error
	if err != nil {
		logrus.Errorf("save runtime Log error: %s", err.Error())
		return
	}
	// 清空 itemList 以便下次保存增加新内容
	//l.itemList = []string{}
}

func (l *RuntimeLog) SetNowTime() {
	l.itemList = append(l.itemList, fmt.Sprintf("<div class=\"log_time\">%s</div>", time.Now().Format("2006-01-02 15:04:05")))
}

func (l *RuntimeLog) SetLevel(level enum.LogLevelType) {
	l.level = level
}

func (l *RuntimeLog) SetTitle(title string) {
	l.title = title
}

func (l *RuntimeLog) setItem(label string, value any, logLevelType enum.LogLevelType) {
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
		logLevelType.String(), label, v))
}

func (l *RuntimeLog) SetItem(label string, value any) {
	l.setItem(label, value, enum.LogInfoLevel)
}

func (l *RuntimeLog) SetItemDebug(label string, value any) {
	l.setItem(label, value, enum.LogDebugLevel)
}

func (l *RuntimeLog) SetItemTrace(label string, value any) {
	l.setItem(label, value, enum.LogTraceLevel)
}

func (l *RuntimeLog) SetItemInfo(label string, value any) {
	l.setItem(label, value, enum.LogInfoLevel)
}

func (l *RuntimeLog) SetItemWarn(label string, value any) {
	l.setItem(label, value, enum.LogWarnLevel)
}

func (l *RuntimeLog) SetItemError(label string, value any) {
	l.setItem(label, value, enum.LogErrorLevel)
}

func (l *RuntimeLog) SetItemFatal(label string, value any) {
	l.setItem(label, value, enum.LogFatalLevel)
}

func (l *RuntimeLog) SetItemPanic(label string, value any) {
	l.setItem(label, value, enum.LogPanicLevel)
}

func (l *RuntimeLog) SetImage(src string) {
	l.itemList = append(l.itemList, fmt.Sprintf("<div class=\"log_image\"><img src=\"%s\" alt=\"\"></div>",
		src))
}

func (l *RuntimeLog) SetLink(label string, href string) {
	l.itemList = append(l.itemList, fmt.Sprintf("<div class=\"log_item link\"><div class=\"log_item_label\">%s</div><div class=\"log_item_content\"><a href=\"%s\" target=\"_blank\">%s</a></div></div> ",
		label, href, href))
}

func (l *RuntimeLog) SetError(label string, err error) {
	msg := e.WithStack(err)
	logrus.Errorf("%s: %s", label, err.Error())
	l.itemList = append(l.itemList, fmt.Sprintf("<div class=\"log_error\"><div class=\"line\"><div class=\"label\">%s</div><div class=\"value\">%s</div><div class=\"type\">%T</div></div><div class=\"stack\">%+v</div></div>",
		label, err, err, msg))
}
