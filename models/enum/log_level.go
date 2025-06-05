// Path: ./models/enum/log_level.go

package enum

type LogLevelType uint8

const (
	LogDebugLevel LogLevelType = 1
	LogTraceLevel LogLevelType = 2
	LogInfoLevel  LogLevelType = 3
	LogWarnLevel  LogLevelType = 4
	LogErrorLevel LogLevelType = 5
	LogFatalLevel LogLevelType = 6
	LogPanicLevel LogLevelType = 7
)

func (l LogLevelType) String() string {
	switch l {
	case LogDebugLevel:
		return "Debug"
	case LogTraceLevel:
		return "Trace"
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
