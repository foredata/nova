package metrics

import "fmt"

// Writer 提供各种指标接口，用于写入最终平台
type Writer interface {
	Write(m Metric)
}

// NewConsoleWriter 控制台输出
func NewConsoleWriter() Writer {
	return &consoleWriter{}
}

type consoleWriter struct {
}

func (w *consoleWriter) Write(m Metric) {
	switch m := m.(type) {
	case Counter:
		fmt.Printf("counter %s, value:%d\n", m.Desc().Name, m.Value())
	case Gauge:
		fmt.Printf("gauge %s, value:%d\n", m.Desc().Name, m.Value())
	case GaugeFloat:
		fmt.Printf("gauge %s, value:%f\n", m.Desc().Name, m.Value())
	}
}
