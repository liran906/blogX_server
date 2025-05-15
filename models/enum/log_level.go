// Path: ./models/enum/log_level.go

package enum

type LogLevelType uint8

const (
	LogDebugLevel LogLevelType = 1
	LogInfoLevel  LogLevelType = 2
	LogWarnLevel  LogLevelType = 3
	LogErrorLevel LogLevelType = 4
	LogFatalLevel LogLevelType = 5
)
