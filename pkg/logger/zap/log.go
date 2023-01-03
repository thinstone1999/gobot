package zap

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	sugar *zap.SugaredLogger // logger
	opt   Options
}

func NewLogger(ops ...Option) *Logger {
	l := &Logger{}
	l.opt.Name = defalutLogName()
	l.opt.Level = zap.DebugLevel
	l.opt.Apply(ops...)
	l.init()
	return l
}

func (l *Logger) Debug(msg string, keysAndValues ...interface{}) {
	l.sugar.Debugw(msg, keysAndValues...)
}

func (l *Logger) Warn(msg string, keysAndValues ...interface{}) {
	l.sugar.Warnw(msg, keysAndValues...)
}

func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	l.sugar.Infow(msg, keysAndValues...)
}

func (l *Logger) Error(msg string, keysAndValues ...interface{}) {
	l.sugar.Errorw(msg, keysAndValues...)
}

func (l *Logger) init() {
	core := l.getCore()
	l.sugar = zap.New(core,
		zap.AddStacktrace(zap.ErrorLevel),
		zap.AddCaller(),
		zap.AddCallerSkip(2),
	).Sugar()

}
func (lg *Logger) getCore() zapcore.Core {
	enc := getZapEncoder()
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   lg.opt.getFname(),
		MaxSize:    5,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   false,
	})

	var wts = []zapcore.WriteSyncer{w}
	if lg.opt.Stdout {
		wts = append(wts, zapcore.AddSync(os.Stdout))
	}

	sync := zapcore.NewMultiWriteSyncer(wts...)
	if !lg.opt.ErrorAlone {
		return zapcore.NewCore(enc, sync, lg.opt.Level)
	}

	errorFn := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return l > zapcore.InfoLevel
	})

	infoFn := zap.LevelEnablerFunc(func(l zapcore.Level) bool {

		return l >= lg.opt.Level && l <= zapcore.InfoLevel
	})

	esync := zapcore.AddSync(&lumberjack.Logger{
		Filename:   lg.opt.getErrorname(),
		MaxSize:    1,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   false,
	})
	return zapcore.NewTee(
		zapcore.NewCore(enc, sync, infoFn),
		zapcore.NewCore(enc, esync, errorFn),
	)
}

func getZapEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, pae zapcore.PrimitiveArrayEncoder) {
		pae.AppendString(t.Local().Format("2006-01-02 15:04:05.000"))
	}

	encoderConfig.EncodeCaller = func(ec zapcore.EntryCaller, pae zapcore.PrimitiveArrayEncoder) {
		pae.AppendString(ec.TrimmedPath())
	}

	encoderConfig.CallerKey = "caller_line"
	encoderConfig.EncodeLevel = func(l zapcore.Level, pae zapcore.PrimitiveArrayEncoder) {
		pae.AppendString("[" + l.CapitalString() + "]")
	}

	return zapcore.NewConsoleEncoder(encoderConfig)
}

func defalutLogName() string {
	os.Getpid()
	exe := filepath.Base(os.Args[0])
	name := strings.Split(exe, filepath.Ext(exe))[0]
	return name
}
