package metrics

import (
	"math"
	"sync/atomic"
)

// GaugeFloat like Gauge, but type type is float64
type GaugeFloat interface {
	Metric
	Value() float64
	Set(float64)
	Add(float64)
	Sub(float64)
}

type GaugeFloatSet interface {
	Labels(labels Labels) GaugeFloat
	Values(labels ...string) GaugeFloat
}

type GaugeFloatOpts = Opts

func newGaugeFloat(desc *Desc) GaugeFloat {
	g := &gaugeFloat{desc: desc}
	return g
}

type gaugeFloat struct {
	value uint64
	desc  *Desc
}

func (g *gaugeFloat) Desc() *Desc {
	return g.desc
}

func (g *gaugeFloat) Value() float64 {
	return math.Float64frombits(atomic.LoadUint64(&g.value))
}

func (g *gaugeFloat) Set(v float64) {
	atomic.StoreUint64(&g.value, math.Float64bits(v))
}

func (g *gaugeFloat) Add(v float64) {
	for {
		oldBits := atomic.LoadUint64(&g.value)
		newBits := math.Float64bits(math.Float64frombits(oldBits) + v)
		if atomic.CompareAndSwapUint64(&g.value, oldBits, newBits) {
			return
		}
	}
}

func (g *gaugeFloat) Sub(v float64) {
	g.Add(v * -1)
}

func newGaugeFloatSet(r Registry, opts *GaugeFloatOpts, labelNames []string) GaugeFloatSet {
	s := &gaugeFloatSet{}
	s.init(r, toDesc(opts), labelNames, func(desc *Desc) Metric {
		return newGaugeFloat(desc)
	})

	return s
}

type gaugeFloatSet struct {
	metricSet
}

func (s *gaugeFloatSet) Labels(labels Labels) GaugeFloat {
	x := s.getOrCreateByLabels(labels)
	if x != nil {
		return x.(GaugeFloat)
	}
	return nil
}

func (s *gaugeFloatSet) Values(labels ...string) GaugeFloat {
	x := s.getOrCreateByValues(labels...)
	if x != nil {
		return x.(GaugeFloat)
	}
	return nil
}
