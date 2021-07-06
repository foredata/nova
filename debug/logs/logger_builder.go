package logs

import (
	"context"
	"fmt"
	"sync"
)

var gBuilderPool = sync.Pool{
	New: func() interface{} {
		return &Builder{}
	},
}

func newBuilder(e *Entry) *Builder {
	b := gBuilderPool.New().(*Builder)
	b.entry = e
	return b
}

// Builder 可通过With等方法构建Logger
type Builder struct {
	entry *Entry
}

func (b *Builder) free() {
	gBuilderPool.Put(b)
}

func (b *Builder) WithCtx(ctx context.Context) *Builder {
	b.entry.Context = ctx
	return b
}

func (b *Builder) WithFields(fields ...Field) *Builder {
	b.entry.Fields = append(b.entry.Fields, fields...)
	return b
}

func (b *Builder) WithDepth(depth int) *Builder {
	b.entry.CallDepth = depth
	return b
}

func (b *Builder) Log(lv Level, text string, fields ...Field) {
	e := b.entry
	l := e.Logger
	if l.IsEnable(lv) {
		e := b.entry
		e.Level = lv
		e.Text = text
		e.Fields = fields
		l.Write(e)
	}
	b.free()
}

func (b *Builder) Logf(lv Level, format string, args ...interface{}) {
	e := b.entry
	l := e.Logger

	if l.IsEnable(lv) {
		text := fmt.Sprintf(format, args...)
		e.Level = lv
		e.Text = text
		l.Write(e)
	}
	b.free()
}

func (b *Builder) Trace(msg string, fields ...Field) {
	b.Log(TraceLevel, msg, fields...)
}

func (b *Builder) Debug(msg string, fields ...Field) {
	b.Log(DebugLevel, msg, fields...)
}

func (b *Builder) Info(msg string, fields ...Field) {
	b.Log(InfoLevel, msg, fields...)
}

func (b *Builder) Warn(msg string, fields ...Field) {
	b.Log(WarnLevel, msg, fields...)
}

func (b *Builder) Error(msg string, fields ...Field) {
	b.Log(ErrorLevel, msg, fields...)
}

func (b *Builder) Fatal(msg string, fields ...Field) {
	b.Log(FatalLevel, msg, fields...)
}

func (b *Builder) Tracef(format string, args ...interface{}) {
	b.Logf(TraceLevel, format, args...)
}

func (b *Builder) Debugf(format string, args ...interface{}) {
	b.Logf(DebugLevel, format, args...)
}

func (b *Builder) Infof(format string, args ...interface{}) {
	b.Logf(InfoLevel, format, args...)
}

func (b *Builder) Warnf(format string, args ...interface{}) {
	b.Logf(WarnLevel, format, args...)
}

func (b *Builder) Errorf(format string, args ...interface{}) {
	b.Logf(ErrorLevel, format, args...)
}

func (b *Builder) Fatalf(format string, args ...interface{}) {
	b.Logf(FatalLevel, format, args...)
}

func (b *Builder) Tracew(format string, args ...interface{}) {
	b.Logf(TraceLevel, format, args...)
}
