package debug

var gDefault Debuger = &noopDebuger{}

func SetDefault(d Debuger) {
	if d != nil {
		gDefault = d
	} else {
		gDefault = &noopDebuger{}
	}
}

func Default() Debuger {
	return gDefault
}

func Errorf(format string, args ...interface{}) {
	gDefault.Errorf(format, args...)
}

func Infof(format string, args ...interface{}) {
	gDefault.Infof(format, args...)
}

func Tracef(format string, args ...interface{}) {
	gDefault.Tracef(format, args...)
}

// Debuger 内部调试使用
type Debuger interface {
	Errorf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Tracef(format string, args ...interface{})
}

type noopDebuger struct {
}

func (*noopDebuger) Errorf(format string, args ...interface{}) {
}

func (*noopDebuger) Infof(format string, args ...interface{}) {
}

func (*noopDebuger) Tracef(format string, args ...interface{}) {
}
