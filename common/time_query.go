// Path: ./common/time_query.go

package common

import (
	"blogX_server/global"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"time"
)

// TimeQuery 按照开始时间、结束时间
func TimeQuery(start, end string) (query *gorm.DB, err error) {
	query = global.DB.Where("")
	if start != "" {
		_, err = time.Parse("2006-01-02 15:04:05", start)
		if err != nil {
			err = errors.New(fmt.Sprintf("开始时间[%s]格式错误: %s", start, err.Error()))
			return
		}
		query = query.Where("created_at >= ?", start)
	}
	if end != "" {

		_, err = time.Parse("2006-01-02 15:04:05", end)
		if err != nil {
			err = errors.New(fmt.Sprintf("结束时间[%s]格式错误: %s", end, err.Error()))
			return
		}
		if start != "" && start >= end {
			err = errors.New("开始时间必须早于结束时间")
			return
		}
		query = query.Where("created_at <= ?", end)
	}
	return
}
