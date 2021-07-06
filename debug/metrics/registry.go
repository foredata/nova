package metrics

import (
	"sync"
)

var (
	defaultRegistry = NewRegistry()
)

// SetDefault 设置默认Registry
func SetDefault(r Registry) {
	defaultRegistry = r
}

// NewRegistry 新建Registry
func NewRegistry() Registry {
	r := &registry{}
	return r
}

// NewCounter create counter
func NewCounter(opts *CounterOpts) Counter {
	return defaultRegistry.NewCounter(opts)
}

// NewGauge create gauge
func NewGauge(opts *Opts) Gauge {
	return defaultRegistry.NewGauge(opts)
}

// NewGaugeFloat create gauge
func NewGaugeFloat(opts *GaugeFloatOpts) GaugeFloat {
	return defaultRegistry.NewGaugeFloat(opts)
}

func NewMeter(opts *MeterOpts) Meter {
	return defaultRegistry.NewMeter(opts)
}

func NewHistogram(opts *HistogramOpts) Histogram {
	return defaultRegistry.NewHistogram(opts)
}

func NewSummary(opts *SummaryOpts) Summary {
	return defaultRegistry.NewSummary(opts)
}

func NewTimer(opts *TimerOpts) Timer {
	return defaultRegistry.NewTimer(opts)
}

func NewCounterSet(opts *CounterOpts, labelNames []string) CounterSet {
	return defaultRegistry.NewCounterSet(opts, labelNames)
}
func NewGaugeSet(opts *GaugeOpts, labelNames []string) GaugeSet {
	return defaultRegistry.NewGaugeSet(opts, labelNames)
}

func NewGaugeFloatSet(opts *GaugeFloatOpts, labelNames []string) GaugeFloatSet {
	return defaultRegistry.NewGaugeFloatSet(opts, labelNames)
}

func NewMeterSet(opts *MeterOpts, labelNames []string) MeterSet {
	return defaultRegistry.NewMeterSet(opts, labelNames)
}

func NewHistogramSet(opts *HistogramOpts, labelNames []string) HistogramSet {
	return defaultRegistry.NewHistogramSet(opts, labelNames)
}

func NewSummarySet(opts *SummaryOpts, labelNames []string) SummarySet {
	return defaultRegistry.NewSummarySet(opts, labelNames)
}

func NewTimerSet(opts *TimerOpts, labelNames []string) TimerSet {
	return defaultRegistry.NewTimerSet(opts, labelNames)
}

// Flush .
func Flush(w Writer) {
	defaultRegistry.Flush(w)
}

// Registry 创建Metric并注册
type Registry interface {
	NewCounter(opts *CounterOpts) Counter
	NewGauge(opts *GaugeOpts) Gauge
	NewGaugeFloat(opts *GaugeFloatOpts) GaugeFloat
	NewMeter(opts *MeterOpts) Meter
	NewHistogram(opts *HistogramOpts) Histogram
	NewSummary(opts *SummaryOpts) Summary
	NewTimer(opts *TimerOpts) Timer

	NewCounterSet(opts *CounterOpts, labelNames []string) CounterSet
	NewGaugeSet(opts *GaugeOpts, labelNames []string) GaugeSet
	NewGaugeFloatSet(opts *GaugeFloatOpts, labelNames []string) GaugeFloatSet
	NewMeterSet(opts *MeterOpts, labelNames []string) MeterSet
	NewHistogramSet(opts *HistogramOpts, labelNames []string) HistogramSet
	NewSummarySet(opts *SummaryOpts, labelNames []string) SummarySet
	NewTimerSet(opts *TimerOpts, labelNames []string) TimerSet

	Register(m Metric)
	Flush(w Writer)
}

type registry struct {
	metrics []Metric
	mux     sync.RWMutex
}

func (r *registry) NewCounter(opts *CounterOpts) Counter {
	c := newCounter(toDesc(opts))
	r.Register(c)

	return c
}

func (r *registry) NewGauge(opts *GaugeOpts) Gauge {
	g := newGauge(toDesc(opts))
	r.Register(g)
	return g
}

func (r *registry) NewGaugeFloat(opts *GaugeFloatOpts) GaugeFloat {
	g := newGaugeFloat(toDesc(opts))
	r.Register(g)

	return g
}

func (r *registry) NewMeter(opts *MeterOpts) Meter {
	m := newMeter(toDesc(opts))
	r.Register(m)
	return m
}

func (r *registry) NewHistogram(opts *HistogramOpts) Histogram {
	desc := &Desc{
		Name:   toFullName(opts.Namespace, opts.Subsystem, opts.Name),
		Help:   opts.Help,
		Labels: toLabels(opts.Labels),
	}

	h := newHistogram(desc)
	r.Register(h)
	return h
}

func (r *registry) NewSummary(opts *SummaryOpts) Summary {
	desc := &Desc{
		Name:   toFullName(opts.Namespace, opts.Subsystem, opts.Name),
		Help:   opts.Help,
		Labels: toLabels(opts.Labels),
	}

	s := newSummary(desc)
	r.Register(s)
	return s
}

func (r *registry) NewTimer(opts *TimerOpts) Timer {
	t := newTimer(toDesc(opts))
	r.Register(t)
	return t
}

func (r *registry) NewCounterSet(opts *CounterOpts, labelNames []string) CounterSet {
	return newCounterSet(r, opts, labelNames)
}

func (r *registry) NewGaugeSet(opts *GaugeOpts, labelNames []string) GaugeSet {
	return newGaugeSet(r, opts, labelNames)
}

func (r *registry) NewGaugeFloatSet(opts *GaugeFloatOpts, labelNames []string) GaugeFloatSet {
	return newGaugeFloatSet(r, opts, labelNames)
}

func (r *registry) NewMeterSet(opts *MeterOpts, labelNames []string) MeterSet {
	return nil
}

func (r *registry) NewHistogramSet(opts *HistogramOpts, labelNames []string) HistogramSet {
	return newHistogramSet(r, opts, labelNames)
}

func (r *registry) NewSummarySet(opts *SummaryOpts, labelNames []string) SummarySet {
	return newSummarySet(r, opts, labelNames)
}

func (r *registry) NewTimerSet(opts *TimerOpts, labelNames []string) TimerSet {
	return nil
}

func (r *registry) Register(m Metric) {
	r.mux.Lock()
	r.metrics = append(r.metrics, m)
	r.mux.Unlock()
}

func (r *registry) Flush(w Writer) {
	r.mux.Lock()
	for _, m := range r.metrics {
		w.Write(m)
	}
	r.mux.Unlock()
}
