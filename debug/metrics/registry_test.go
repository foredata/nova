package metrics_test

import (
	"testing"

	"github.com/foredata/nova/debug/metrics"
)

func TestCounter(t *testing.T) {
	c := metrics.NewCounter(&metrics.CounterOpts{
		Name:   "demo",
		Help:   "test",
		Labels: metrics.Labels{"ip": "127.0.0.1"},
	})

	for i := 0; i < 10; i++ {
		c.Inc()
	}
	w := metrics.NewConsoleWriter()
	metrics.Flush(w)
}
