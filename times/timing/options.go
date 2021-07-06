package timing

// Options 创建Timer可选参数
type Options struct {
	engine Engine
}

type Option func(o *Options)

func newOptions(opts ...Option) *Options {
	o := &Options{engine: gDefault}
	for _, fn := range opts {
		fn(o)
	}

	return o
}

func WithEngine(e Engine) Option {
	return func(o *Options) {
		if e != nil {
			o.engine = e
		}
	}
}
