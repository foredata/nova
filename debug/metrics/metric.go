package metrics

import (
	"errors"
	"sort"
	"strings"
	"sync"
)

var (
	errLabelSizeNotMatch  = errors.New("metrics: label size not match")
	errLabelNotFoundValue = errors.New("metrics: not found value from labels")
)

// Pair key-value pair
type Pair struct {
	Name  string
	Value string
}

// Labels represents a collection of label name -> value mappings. This type is
// commonly used with the With(Labels) and GetMetricWith(Labels) methods of
// metric vector Collectors, e.g.:
//     myVec.With(Labels{"code": "404", "method": "GET"}).Add(42)
//
// The other use-case is the specification of constant label pairs in Opts or to
// create a Desc.
type Labels map[string]string

// Desc is the descriptor used by every Metric. It is essentially
// the immutable meta-data of a Metric. The normal Metric implementations
// included in this package manage their Desc under the hood. Users only have to
// deal with Desc if they use advanced features like the ExpvarCollector or
// custom Collectors and Metrics.
//
// Descriptors registered with the same registry have to fulfill certain
// consistency and uniqueness criteria if they share the same fully-qualified
// name: They must have the same help string and the same label names (aka label
// dimensions) in each, constLabels and variableLabels, but they must differ in
// the values of the constLabels.
//
// Descriptors that share the same fully-qualified names and the same label
// values of their constLabels are considered equal.
//
// Use NewDesc to create new Desc instances.
type Desc struct {
	Name   string //
	Help   string //
	Labels []Pair // 所有标签,包含静态标签和动态标签,静态标签在前边,动态标签在后边
}

// clone 赋值新Desc，并添加动态标签
func (d *Desc) clone(labelNames, labelValues []string) *Desc {
	c := &Desc{
		Name:   d.Name,
		Help:   d.Help,
		Labels: d.Labels,
	}

	for i, n := range labelNames {
		c.Labels = append(c.Labels, Pair{Name: n, Value: labelValues[i]})
	}

	return c
}

type Opts struct {
	// Namespace, Subsystem, and Name are components of the fully-qualified
	// name of the Metric (created by joining these components with
	// "_"). Only Name is mandatory, the others merely help structuring the
	// name. Note that the fully-qualified name of the metric must be a
	// valid Prometheus metric name.
	Namespace string
	Subsystem string
	Name      string
	// Help provides information about this metric.
	//
	// Metrics with the same fully-qualified name must have the same Help string.
	Help   string
	Labels Labels
}

func toDesc(opts *Opts) *Desc {
	desc := &Desc{
		Name:   toFullName(opts.Namespace, opts.Subsystem, opts.Name),
		Help:   opts.Help,
		Labels: toLabels(opts.Labels),
	}

	return desc
}

func toFullName(namespace, subsystem, name string) string {
	if name == "" {
		return ""
	}

	switch {
	case namespace != "" && subsystem != "":
		return strings.Join([]string{namespace, subsystem, name}, "_")
	case namespace != "":
		return strings.Join([]string{namespace, name}, "_")
	case subsystem != "":
		return strings.Join([]string{subsystem, name}, "_")
	}
	return name
}

func toLabels(labels Labels) []Pair {
	r := make([]Pair, 0, len(labels))
	for k, v := range labels {
		r = append(r, Pair{Name: k, Value: v})
	}

	sort.Slice(r, func(i, j int) bool {
		return r[i].Name < r[j].Name
	})

	return r
}

// A Metric models a single sample value with its meta data being exported to
// Prometheus. Implementations of Metric in this package are Gauge, Counter,
// Histogram, Summary, and Untyped.
type Metric interface {
	// Desc returns the descriptor for the Metric. This method idempotently
	// returns the same descriptor throughout the lifetime of the
	// Metric. The returned descriptor is immutable by contract. A Metric
	// unable to describe itself must return an invalid descriptor (created
	// with NewInvalidDesc).
	Desc() *Desc
}

type factory func(desc *Desc) Metric

// metricSet is a Collector to bundle metrics of the same name that differ in
// their label values. MetricVec is not used directly but as a building block
// for implementations of vectors of a given metric type, like GaugeSet,
// CounterSet, SummarySet, and HistogramSet. It is exported so that it can be
// used for custom Metric implementations.
type metricSet struct {
	mux        sync.RWMutex        //
	reg        Registry            //
	creator    factory             //
	desc       *Desc               //
	labelNames []string            // 动态标签名
	metrics    map[uint64][]Metric //
}

func (s *metricSet) init(reg Registry, desc *Desc, labelNames []string, creator factory) {
	s.reg = reg
	s.desc = desc
	s.labelNames = labelNames
	s.creator = creator
}

// getOrCreateByLabels 通过标签查找Metric,如果标签与注册时不一致,则返回空
// 性能会比getOrCreateByValues低一些
func (s *metricSet) getOrCreateByLabels(labels Labels) Metric {
	values := make([]string, 0, len(s.labelNames))
	for _, key := range s.labelNames {
		val, ok := labels[key]
		if !ok {
			return nil
		}
		values = append(values, val)
	}

	return s.getOrCreateByValues(values...)
}

// getOrCreateByValues 通过label values查找Metric
func (s *metricSet) getOrCreateByValues(values ...string) Metric {
	if len(values) != len(s.labelNames) {
		return nil
	}

	h := hashNew()
	for _, v := range values {
		h = hashAdd(h, v)
	}

	s.mux.RLock()
	m := s.findByValues(h, values)
	s.mux.RUnlock()
	if m != nil {
		return m
	}

	s.mux.Lock()
	m = s.findByValues(h, values)
	if m == nil {
		m = s.createByValues(h, values)
	}
	s.mux.Unlock()

	return m
}

func (s *metricSet) findByValues(hash uint64, values []string) Metric {
	arr := s.metrics[hash]
	if len(arr) == 0 {
		return nil
	}

	labelStart := len(s.desc.Labels)

	for _, m := range arr {
		eq := true
		labels := m.Desc().Labels[labelStart:]
		for i := 0; i < len(labels); i++ {
			if labels[i].Value != values[i] {
				eq = false
				break
			}
		}

		if eq {
			return m
		}
	}

	return nil
}

func (s *metricSet) createByValues(hash uint64, values []string) Metric {
	desc := s.desc.clone(s.labelNames, values)
	m := s.creator(desc)
	s.reg.Register(m)

	arr := s.metrics[hash]
	s.metrics[hash] = append(arr, m)
	return m
}
