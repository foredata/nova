package logs

import "context"

var defaultLogger = NewDefault()

// SetDefault 设置默认的log
func SetDefault(l Logger) {
	if defaultLogger != nil {
		defaultLogger.Stop()
	}
	defaultLogger = l
}

// NewDefault 创建默认的Logger,默认只包含Console的输出通路
func NewDefault() Logger {
	conf := NewConfig()
	conf.Channels = append(conf.Channels, NewConsoleChannel())
	return NewLogger(conf)
}

func WithFields(fields ...Field) *Builder {
	e := NewEntry(defaultLogger)
	b := newBuilder(e)
	return b
}

func WithCtx(ctx context.Context) *Builder {
	e := NewEntry(defaultLogger)
	b := newBuilder(e)
	return b
}

func Trace(msg string, fields ...Field) {
	defaultLogger.Log(TraceLevel, msg, fields...)
}

func Debug(msg string, fields ...Field) {
	defaultLogger.Log(DebugLevel, msg, fields...)
}

func Info(msg string, fields ...Field) {
	defaultLogger.Log(InfoLevel, msg, fields...)
}

func Warn(msg string, fields ...Field) {
	defaultLogger.Log(WarnLevel, msg, fields...)
}

func Error(msg string, fields ...Field) {
	defaultLogger.Log(ErrorLevel, msg, fields...)
}

func Fatal(msg string, fields ...Field) {
	defaultLogger.Log(FatalLevel, msg, fields...)
}

func Tracef(format string, args ...interface{}) {
	defaultLogger.Logf(TraceLevel, format, args...)
}

func Debugf(format string, args ...interface{}) {
	defaultLogger.Logf(DebugLevel, format, args...)
}

func Infof(format string, args ...interface{}) {
	defaultLogger.Logf(InfoLevel, format, args...)
}

func Warnf(format string, args ...interface{}) {
	defaultLogger.Logf(WarnLevel, format, args...)
}

func Errorf(format string, args ...interface{}) {
	defaultLogger.Logf(ErrorLevel, format, args...)
}

func Fatalf(format string, args ...interface{}) {
	defaultLogger.Logf(FatalLevel, format, args...)
}
