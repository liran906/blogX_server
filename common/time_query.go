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
	return TimeQueryWithBase(global.DB.Where(""), start, end)
}

// TimeQueryWithBase 接受初始查询条件
func TimeQueryWithBase(query *gorm.DB, start, end string) (*gorm.DB, error) {
	layout := "2006-01-02 15:04:05"
	if start != "" {
		_, err := time.Parse(layout, start)
		if err != nil {
			err = errors.New(fmt.Sprintf("开始时间[%s]格式错误: %s", start, err.Error()))
			return nil, err
		}
		query = query.Where("created_at >= ?", start)
	}

	if end != "" {
		_, err := time.Parse(layout, end)
		if err != nil {
			err = errors.New(fmt.Sprintf("结束时间[%s]格式错误: %s", end, err.Error()))
			return nil, err
		}
		if start != "" && start >= end {
			err = errors.New("开始时间必须早于结束时间")
			return nil, err
		}
		query = query.Where("created_at <= ?", end)
	}
	return query, nil
}
