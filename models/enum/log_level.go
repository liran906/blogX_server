// Path: ./models/enum/log_level.go

package enum

type LogLevelType uint8

const (
	LogDebugLevel LogLevelType = 1
	LogInfoLevel  LogLevelType = 2
	LogWarnLevel  LogLevelType = 3
	LogErrorLevel LogLevelType = 4
	LogFatalLevel LogLevelType = 5
	LogPanicLevel LogLevelType = 6
)

func (l LogLevelType) ToString() string {
	switch l {
	case LogDebugLevel:
		return "Debug"
	case LogInfoLevel:
		return "Info"
	case LogWarnLevel:
		return "Warn"
	case LogErrorLevel:
		return "Error"
	case LogFatalLevel:
		return "Fatal"
	case LogPanicLevel:
		return "Panic"
	}
	return ""
}
