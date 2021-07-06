package metrics

type MeterModel struct {
}

// Meter 度量某个时间段的平均处理次数（request  per second）
type Meter interface {
	Metric
	Value() *MeterModel
	Add(n int64)
}

type MeterSet interface {
	Labels(labels Labels) Meter
	Values(labels ...string) Meter
}

type MeterOpts = Opts

func newMeter(desc *Desc) Meter {
	m := &meter{desc: desc}
	return m
}

// TODO: implementation
type meter struct {
	desc *Desc
}

func (m *meter) Desc() *Desc {
	return m.desc
}

func (m *meter) Value() *MeterModel {
	return nil
}

func (m *meter) Write(w Writer) {
}

func (m *meter) Add(v int64) {
}

func newMeterSet(r Registry, opts *GaugeFloatOpts, labelNames []string) MeterSet {
	s := &meterSet{}
	s.init(r, toDesc(opts), labelNames, func(desc *Desc) Metric {
		return newMeter(desc)
	})

	return s
}

type meterSet struct {
	metricSet
}

func (s *meterSet) Labels(labels Labels) Meter {
	x := s.getOrCreateByLabels(labels)
	if x != nil {
		return x.(Meter)
	}
	return nil
}

func (s *meterSet) Values(labels ...string) Meter {
	x := s.getOrCreateByValues(labels...)
	if x != nil {
		return x.(Meter)
	}
	return nil
}
