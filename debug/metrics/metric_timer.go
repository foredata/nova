package metrics

type TimerModel struct {
}

// Timer 是Histogram和Meter的结合，Histogram统计耗时分布，Meter统计QPS；
type Timer interface {
	Metric
	Value() *TimerModel
}

type TimerSet interface {
	Labels(labels Labels) Timer
	Values(labels ...string) Timer
}

type TimerOpts = Opts

func newTimer(desc *Desc) Timer {
	t := &timer{desc: desc}
	return t
}

// TODO: implementation
type timer struct {
	desc *Desc
}

func (t *timer) Desc() *Desc {
	return t.desc
}

func (t *timer) Value() *TimerModel {
	return nil
}

func newTimerSet(r Registry, opts *GaugeFloatOpts, labelNames []string) TimerSet {
	s := &timerSet{}
	s.init(r, toDesc(opts), labelNames, func(desc *Desc) Metric {
		return newTimer(desc)
	})

	return s
}

type timerSet struct {
	metricSet
}

func (s *timerSet) Labels(labels Labels) Timer {
	x := s.getOrCreateByLabels(labels)
	if x != nil {
		return x.(Timer)
	}
	return nil
}

func (s *timerSet) Values(labels ...string) Timer {
	x := s.getOrCreateByValues(labels...)
	if x != nil {
		return x.(Timer)
	}
	return nil
}
