// Path: ./service/log_service/runtime_log.go

package log_service

import (
	"blogX_server/global"
	"blogX_server/models"
	"blogX_server/models/enum"
	"github.com/sirupsen/logrus"
	"strings"
)

type RuntimeLog struct {
	title        string
	content      string
	level        enum.LogLevelType
	itemList     []string
	log          *models.LogModel
	serviceName  string
	runtimeDelta RuntimeDelta
}

type RuntimeDelta int8

const (
	RuntimeDeltaHour  RuntimeDelta = 1
	RuntimeDeltaDay   RuntimeDelta = 2
	RuntimeDeltaWeek  RuntimeDelta = 3
	RuntimeDeltaMonth RuntimeDelta = 4
)

func (l *RuntimeLog) NewRuntimeLog(serviceName string, delta RuntimeDelta) *RuntimeLog {
	return &RuntimeLog{
		title:        l.title,
		runtimeDelta: delta,
	}
}

func (l *RuntimeLog) Save() {
	content := strings.Join(l.itemList, "\n")
	log := models.LogModel{
		Title:   l.title,
		Content: content,
		Level:   l.level,
		LogType: enum.RuntimeLogType,
	}
	err := global.DB.Create(log).Error
	if err != nil {
		logrus.Errorf("save runtime Log error: %s", err.Error())
		return
	}
	l.log = &log
}
