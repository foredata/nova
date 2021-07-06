package metrics

import (
	"sync/atomic"
)

// Counter is a Metric that represents a single numerical value that only ever
// goes up. That implies that it cannot be used to count items whose number can
// also go down, e.g. the number of currently running goroutines. Those
// "counters" are represented by Gauges.
//
// A Counter is typically used to count requests served, tasks completed, errors
// occurred, etc.
//
// To create Counter instances, use NewCounter.
type Counter interface {
	Metric
	Value() int64
	Add(int64)
	Inc()
	Dec()
}

// CounterSet is a Collector that bundles a set of Counters that all share the
// same Desc, but have different values for their variable labels. This is used
// if you want to count the same thing partitioned by various dimensions
// (e.g. number of HTTP requests, partitioned by response code and
// method). Create instances with CounterSet.
type CounterSet interface {
	Labels(labels Labels) Counter
	Values(labels ...string) Counter
}

type CounterOpts = Opts

func newCounter(desc *Desc) Counter {
	c := &counter{desc: desc}
	return c
}

type counter struct {
	value int64
	desc  *Desc
}

func (c *counter) Desc() *Desc {
	return c.desc
}

func (c *counter) Value() int64 {
	return atomic.LoadInt64(&c.value)
}

func (c *counter) Add(v int64) {
	atomic.AddInt64(&c.value, v)
}

func (c *counter) Inc() {
	atomic.AddInt64(&c.value, 1)
}

func (c *counter) Dec() {
	atomic.AddInt64(&c.value, -1)
}

func newCounterSet(r Registry, opts *GaugeFloatOpts, labelNames []string) CounterSet {
	s := &counterSet{}
	s.init(r, toDesc(opts), labelNames, func(desc *Desc) Metric {
		return newCounter(desc)
	})

	return s
}

type counterSet struct {
	metricSet
}

func (s *counterSet) Labels(labels Labels) Counter {
	x := s.getOrCreateByLabels(labels)
	if x != nil {
		return x.(Counter)
	}
	return nil
}

func (s *counterSet) Values(labels ...string) Counter {
	x := s.getOrCreateByValues(labels...)
	if x != nil {
		return x.(Counter)
	}
	return nil
}
