// Path: ./common/time_transaction.go

package common

import "time"

func DateTimeToUnix(year, month, day, hour, minute, second int) int64 {
	return time.Date(year, time.Month(month), day, hour, minute, second, 0, time.Local).Unix()
}

func UnixToDateTimeString(unix int64) string {
	return time.Unix(unix, 0).Format("2006-01-02 15:04:05")
}
