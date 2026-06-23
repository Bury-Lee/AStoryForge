// core/init_logrus.go
package core

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/sirupsen/logrus"
)

// ANSI color codes used by console logging.
const (
	red    = 31
	yellow = 33
	blue   = 36
	gray   = 34
	cyan   = 36
)

// ConsoleFormatter renders colored output for the terminal.
type ConsoleFormatter struct{}

func (t *ConsoleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var levelColor int
	switch entry.Level {
	case logrus.DebugLevel, logrus.TraceLevel:
		levelColor = cyan
	case logrus.WarnLevel:
		levelColor = yellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = red
	default:
		levelColor = blue
	}

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	timestamp := entry.Time.Format("2006-01-02 15:04:05")

	if entry.HasCaller() {
		funcVal := entry.Caller.Function
		fileVal := fmt.Sprintf("%s:%d", path.Base(entry.Caller.File), entry.Caller.Line)
		// Render caller information with ANSI colors.
		fmt.Fprintf(b, "[%s] \x1b[%dm[%s]\x1b[0m %s %s %s\n",
			timestamp, levelColor, entry.Level, fileVal, funcVal, entry.Message)
	} else {
		fmt.Fprintf(b, "[%s] \x1b[%dm[%s]\x1b[0m %s\n",
			timestamp, levelColor, entry.Level, entry.Message)
	}
	return b.Bytes(), nil
}

// FileFormatter renders plain text log lines for log files.
type FileFormatter struct{}

func (t *FileFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	timestamp := entry.Time.Format("2006-01-02 15:04:05")

	if entry.HasCaller() {
		funcVal := entry.Caller.Function
		fileVal := fmt.Sprintf("%s:%d", path.Base(entry.Caller.File), entry.Caller.Line)
		// Write plain text without ANSI color codes.
		fmt.Fprintf(b, "[%s] [%s] %s %s %s\n",
			timestamp, entry.Level, fileVal, funcVal, entry.Message)
	} else {
		fmt.Fprintf(b, "[%s] [%s] %s\n",
			timestamp, entry.Level, entry.Message)
	}
	return b.Bytes(), nil
}

// FileDateHook rotates log files by date.
type FileDateHook struct {
	file      *os.File
	logPath   string
	fileDate  string
	appName   string
	formatter *FileFormatter
}

func (hook FileDateHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook FileDateHook) Fire(entry *logrus.Entry) error {
	timer := entry.Time.Format("2006-01-02")

	// Use the file formatter so log files do not contain ANSI codes.
	line, err := hook.formatter.Format(entry)
	if err != nil {
		return err
	}

	if hook.fileDate == timer {
		hook.file.Write(line)
		return nil
	}

	// Rotate to a new file when the date changes.
	hook.file.Close()
	os.MkdirAll(fmt.Sprintf("%s/%s", hook.logPath, timer), os.ModePerm)
	filename := fmt.Sprintf("%s/%s/%s.log", hook.logPath, timer, hook.appName)

	hook.file, _ = os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	hook.fileDate = timer
	hook.file.Write(line)
	return nil
}

func InitFile(logPath, appName string) {
	fileDate := time.Now().Format("2006-01-02")
	err := os.MkdirAll(fmt.Sprintf("%s/%s", logPath, fileDate), os.ModePerm)
	if err != nil {
		logrus.Error(err)
		return
	}

	filename := fmt.Sprintf("%s/%s/%s.log", logPath, fileDate, appName)
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		logrus.Error("无法打开日志文件!建议停止程序:", err)
		return
	}

	fileHook := FileDateHook{
		file:      file,
		logPath:   logPath,
		fileDate:  fileDate,
		appName:   appName,
		formatter: &FileFormatter{},
	}
	logrus.AddHook(&fileHook)
}

func InitLogrus() {
	logrus.SetOutput(os.Stdout) // Write logs to stdout.
	// Validate that the log level is configured.
	if Config.Log.LogLevel == "" {
		logrus.Error("日志级别未配置,请检查配置文件")
		panic("日志级别未配置,请检查配置文件")
	}
	if Config.RunMode == "develop" { // Enable caller reporting in develop mode.
		logrus.SetReportCaller(true) // Include function names and line numbers.
	}
	logrus.SetFormatter(&ConsoleFormatter{}) // Use colored terminal formatting.
	level, err := logrus.ParseLevel(Config.Log.LogLevel)
	if err != nil {
		logrus.Error(err)
		panic("日志级别解析失败,请检查配置文件")
	}
	logrus.SetLevel(level) // Set the minimum log level.

	l := Config.Log
	InitFile(l.Dir, l.App)
}
