package enum

type LogType uint8

const (
	LoginLogType   LogType = 1
	ActionLogType  LogType = 2
	RuntimeLogType LogType = 3
)
