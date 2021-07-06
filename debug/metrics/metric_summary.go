package metrics

import "sync"

type Quantile struct {
	Quantile float64
	Value    float64
}

type SummaryModel struct {
	SampleCount uint64
	SampleSum   float64
	Quantiles   []Quantile
}

// A Summary captures individual observations from an event or sample stream and
// summarizes them in a manner similar to traditional summary statistics: 1. sum
// of observations, 2. observation count, 3. rank estimations.
//
// A typical use-case is the observation of request latencies. By default, a
// Summary provides the median, the 90th and the 99th percentile of the latency
// as rank estimations. However, the default behavior will change in the
// upcoming v1.0.0 of the library. There will be no rank estimations at all by
// default. For a sane transition, it is recommended to set the desired rank
// estimations explicitly.
//
// Note that the rank estimations cannot be aggregated in a meaningful way with
// the Prometheus query language (i.e. you cannot average or add them). If you
// need aggregatable quantiles (e.g. you want the 99th percentile latency of all
// queries served across all instances of a service), consider the Histogram
// metric type. See the Prometheus documentation for more details.
//
// To create Summary instances, use NewSummary.
type Summary interface {
	Metric
	Value() *SummaryModel
	// Observe adds a single observation to the summary. Observations are
	// usually positive or zero. Negative observations are accepted but
	// prevent current versions of Prometheus from properly detecting
	// counter resets in the sum of observations. See
	// https://prometheus.io/docs/practices/histograms/#count-and-sum-of-observations
	// for details.
	Observe(float64)
}

type SummarySet interface {
	Labels(labels Labels) Summary
	Values(labels ...string) Summary
}

type SummaryOpts struct {
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

func newSummary(desc *Desc) Summary {
	s := &summary{desc: desc}
	return s
}

// TODO: implementation
type summary struct {
	desc *Desc
	mux  sync.Mutex
}

func (s *summary) Desc() *Desc {
	return s.desc
}

func (s *summary) Value() *SummaryModel {
	return nil
}

func (s *summary) Observe(v float64) {

}

func (s *summary) Write(w Writer) {

}

// newSummarySet create counter set
func newSummarySet(r Registry, opts *SummaryOpts, labelNames []string) SummarySet {
	d := &Desc{
		Name:   toFullName(opts.Namespace, opts.Subsystem, opts.Name),
		Help:   opts.Help,
		Labels: toLabels(opts.Labels),
	}
	s := &summarySet{}
	s.init(r, d, labelNames, func(desc *Desc) Metric {
		return newSummary(desc)
	})

	return s
}

type summarySet struct {
	metricSet
}

func (s *summarySet) Labels(labels Labels) Summary {
	x := s.getOrCreateByLabels(labels)
	if x != nil {
		return x.(Summary)
	}
	return nil
}

func (s *summarySet) Values(labels ...string) Summary {
	x := s.getOrCreateByValues(labels...)
	if x != nil {
		return x.(Summary)
	}
	return nil
}
