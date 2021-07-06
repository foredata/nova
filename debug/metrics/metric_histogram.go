package metrics

type Bucket struct {
	CumulativeCount uint64
	UpperBound      float64
}

type HistogramModel struct {
	SampleCount uint64
	SampleSum   float64
	Buckets     []Bucket
}

// A Histogram counts individual observations from an event or sample stream in
// configurable buckets. Similar to a summary, it also provides a sum of
// observations and an observation count.
//
// On the Prometheus server, quantiles can be calculated from a Histogram using
// the histogram_quantile function in the query language.
//
// Note that Histograms, in contrast to Summaries, can be aggregated with the
// Prometheus query language (see the documentation for detailed
// procedures). However, Histograms require the user to pre-define suitable
// buckets, and they are in general less accurate. The Observe method of a
// Histogram has a very low performance overhead in comparison with the Observe
// method of a Summary.
//
// To create Histogram instances, use NewHistogram.
type Histogram interface {
	Metric
	Value() *HistogramModel
	// Observe adds a single observation to the histogram.
	Observe(float64)
}

type HistogramSet interface {
	Labels(labels Labels) Histogram
	Values(labels ...string) Histogram
}

type HistogramOpts struct {
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

func newHistogram(desc *Desc) Histogram {
	h := &histogram{desc: desc}
	return h
}

// TODO: implementation
type histogram struct {
	desc *Desc
}

func (h *histogram) Desc() *Desc {
	return h.desc
}

func (h *histogram) Value() *HistogramModel {
	return nil
}

func (h *histogram) Observe(v float64) {

}

func newHistogramSet(r Registry, opts *HistogramOpts, labelNames []string) HistogramSet {
	d := &Desc{
		Name:   toFullName(opts.Namespace, opts.Subsystem, opts.Name),
		Help:   opts.Help,
		Labels: toLabels(opts.Labels),
	}
	h := &histogramSet{}
	h.init(r, d, labelNames, func(desc *Desc) Metric {
		return newHistogram(desc)
	})

	return h
}

type histogramSet struct {
	metricSet
}

func (s *histogramSet) Labels(labels Labels) Histogram {
	x := s.getOrCreateByLabels(labels)
	if x != nil {
		return x.(Histogram)
	}
	return nil
}

func (s *histogramSet) Values(labels ...string) Histogram {
	x := s.getOrCreateByValues(labels...)
	if x != nil {
		return x.(Histogram)
	}
	return nil
}
