package logger

import (
	"fmt"

	"github.com/Gonewithmyself/gobot/pkg/logger/zap"

	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

var lg Logger

func init() {
	// lg = zap.NewLogger()
}

func Debug(msg string, keysAndValues ...interface{}) {
	lg.Debug(msg, keysAndValues...)
}

func Info(msg string, keysAndValues ...interface{}) {
	lg.Info(msg, keysAndValues...)
}

func Warn(msg string, keysAndValues ...interface{}) {
	lg.Warn(msg, keysAndValues...)
}

func Error(msg string, keysAndValues ...interface{}) {
	lg.Error(msg, keysAndValues...)
}

func Printf(msg string, args ...interface{}) {
	msg = fmt.Sprintf(msg, args...)
	lg.Debug(msg)
}

func SetLogger(l Logger) {
	lg = l
}

func Init(lv zapcore.Level, path string, name string) {
	lg = zap.NewLogger(
		zap.WithLevel(lv),
		zap.WithPath(path),
		zap.WithName(name),
		// zap.EnableStdout(),
	)
}

func GetLogger() Logger {
	return lg
}
