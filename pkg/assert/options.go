package assert

var defaultOptions = newOptions()

type Options struct {
	failNow bool // 立即结束
	strict  bool // 严格模式,比较类型和值
}

type Option func(o *Options)

func newOptions(opts ...Option) *Options {
	o := &Options{failNow: false, strict: true}
	for _, fn := range opts {
		fn(o)
	}

	return o
}

// WithFailNow .
func WithFailNow() Option {
	return func(o *Options) {
		o.failNow = true
	}
}

// WithLoose .
func WithLoose() Option {
	return func(o *Options) {
		o.strict = false
	}
}

// withOptions 内部使用
func withOptions(other *Options) Option {
	return func(o *Options) {
		*o = *other
	}
}
