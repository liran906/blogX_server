package core

import (
	"blogX_server/global"
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"time"
)

// 颜色
const (
	red    = 31
	yellow = 33
	blue   = 36
	gray   = 37
)

type LogFormatter struct{}

// Format 实现Formatter(entry *logrus.Entry) ([]byte, error)接口
func (t *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 根据日志级别设置终端输出颜色
	var levelColor int
	switch entry.Level {
	case logrus.DebugLevel, logrus.TraceLevel:
		levelColor = gray // 调试和追踪日志：灰色
	case logrus.WarnLevel:
		levelColor = yellow // 警告级别：黄色
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = red // 错误及以上：红色
	default:
		levelColor = blue // 其他信息：蓝色
	}

	// 初始化缓冲区用于构造日志内容
	var b *bytes.Buffer
	if entry.Buffer != nil {
		// 如果 logrus 提供了缓冲区（复用），就直接使用
		b = entry.Buffer
	} else {
		// 否则新建一个缓冲区
		b = &bytes.Buffer{}
	}

	// 设置时间格式
	timestamp := entry.Time.Format("2006-01-02 15:04:05")

	// 如果启用了 caller（调用者）信息
	if entry.HasCaller() {
		funcVal := entry.Caller.Function                                                 // 调用的函数名（全路径）
		fileVal := fmt.Sprintf("%s:%d", path.Base(entry.Caller.File), entry.Caller.Line) // 文件名 + 行号

		// 格式化输出带颜色的日志（含时间、等级、位置、函数名、消息）
		fmt.Fprintf(
			b,
			"[%s] \x1b[%dm[%s]\x1b[0m %s %s %s\n",
			timestamp,
			levelColor,
			entry.Level,
			fileVal,
			funcVal,
			entry.Message,
		)
	} else {
		// 不包含调用者信息时，仅输出时间、等级、消息
		fmt.Fprintf(
			b,
			"[%s] \x1b[%dm[%s]\x1b[0m %s\n",
			timestamp,
			levelColor,
			entry.Level,
			entry.Message,
		)
	}
	// 返回格式化后的日志内容（字节切片）
	return b.Bytes(), nil
}

// FileDateHook 是一个自定义的 logrus Hook，实现按时间自动切换日志文件的功能。
type FileDateHook struct {
	file     *os.File // 当前打开的日志文件
	logPath  string   // 日志目录根路径
	fileDate string   // 当前文件对应的时间（精确到分钟），用于判断是否需要切换文件
	appName  string   // 日志文件名前缀（通常为应用名）
}

// Levels 定义该 Hook 监听哪些级别的日志，这里是监听所有级别
func (hook FileDateHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire 是 Hook 的核心方法，会在每条日志触发时调用
func (hook FileDateHook) Fire(entry *logrus.Entry) error {
	// 获取当前时间（格式：2006-01-02）
	timer := entry.Time.Format("2006-01-02")

	// 将日志条目格式化为字符串
	line, _ := entry.String()

	// ✅ 如果当前日志时间和上次记录的 fileDate 一致，直接写入同一个文件
	if hook.fileDate == timer {
		hook.file.Write([]byte(line))
		return nil
	}

	// ❗时间变化：需要关闭旧文件、创建新目录和文件

	// 关闭旧文件
	hook.file.Close()

	// 创建新目录，例如 logs/2025-05-13/
	os.MkdirAll(fmt.Sprintf("%s/%s", hook.logPath, timer), os.ModePerm)

	// 拼接新的日志文件名，例如 logs/2025-05-13/app.log
	filename := fmt.Sprintf("%s/%s/%s.log", hook.logPath, timer, hook.appName)

	// 打开（或创建）新的日志文件
	hook.file, _ = os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)

	// 更新 fileDate，标记当前已写入该时间段日志
	hook.fileDate = timer

	// 写入新日志
	hook.file.Write([]byte(line))
	return nil
}

func InitFile(logPath, appName string) {
	fileDate := time.Now().Format("2006-01-02")
	//创建目录
	err := os.MkdirAll(fmt.Sprintf("%s/%s", logPath, fileDate), os.ModePerm)
	if err != nil {
		logrus.Error(err)
		return
	}

	filename := fmt.Sprintf("%s/%s/%s.log", logPath, fileDate, appName)
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		logrus.Error(err)
		return
	}
	fileHook := FileDateHook{file, logPath, fileDate, appName}
	logrus.AddHook(&fileHook)
}

func InitLogrus() {
	logrus.SetOutput(os.Stdout)          //设置输出类型
	logrus.SetReportCaller(true)         //开启返回函数名和行号
	logrus.SetFormatter(&LogFormatter{}) //设置自己定义的Formatter
	logrus.SetLevel(logrus.DebugLevel)   //设置最低的Level
	l := global.Config.Logrus
	InitFile(l.Dir, l.App)
}
