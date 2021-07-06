package metrics

import (
	"sync/atomic"
)

// Gauge is a Metric that represents a single numerical value that can
// arbitrarily go up and down.
//
// A Gauge is typically used for measured values like temperatures or current
// memory usage, but also "counts" that can go up and down, like the number of
// running goroutines.
//
// To create Gauge instances, use NewGauge.
type Gauge interface {
	Metric
	Value() int64
	Set(int64)
	Add(int64)
	Sub(int64)
	Inc()
	Dec()
}

type GaugeSet interface {
	Labels(labels Labels) Gauge
	Values(labels ...string) Gauge
}

type GaugeOpts = Opts

func newGauge(desc *Desc) Gauge {
	g := &gauge{desc: desc}
	return g
}

type gauge struct {
	value int64
	desc  *Desc
}

func (g *gauge) Desc() *Desc {
	return g.desc
}

func (g *gauge) Value() int64 {
	v := atomic.LoadInt64(&g.value)
	return v
}

func (g *gauge) Set(v int64) {
	atomic.StoreInt64(&g.value, v)
}

func (g *gauge) Add(v int64) {
	atomic.AddInt64(&g.value, v)
}

func (g *gauge) Sub(v int64) {
	atomic.AddInt64(&g.value, v*-1)
}

func (g *gauge) Inc() {
	atomic.AddInt64(&g.value, 1)
}

func (g *gauge) Dec() {
	atomic.AddInt64(&g.value, -1)
}

func newGaugeSet(r Registry, opts *GaugeFloatOpts, labelNames []string) GaugeSet {
	s := &gaugeSet{}
	s.init(r, toDesc(opts), labelNames, func(desc *Desc) Metric {
		return newGauge(desc)
	})

	return s
}

type gaugeSet struct {
	metricSet
}

func (s *gaugeSet) Labels(labels Labels) Gauge {
	x := s.getOrCreateByLabels(labels)
	if x != nil {
		return x.(Gauge)
	}
	return nil
}

func (s *gaugeSet) Values(labels ...string) Gauge {
	x := s.getOrCreateByValues(labels...)
	if x != nil {
		return x.(Gauge)
	}
	return nil
}
