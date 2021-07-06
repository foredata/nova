package recovery

import "context"

type Options struct {
	err *error
	ctx context.Context
}

type Option func(o *Options)

func WithError(err *error) Option {
	return func(o *Options) {
		o.err = err
	}
}

func WithCtx(ctx context.Context) Option {
	return func(o *Options) {
		o.ctx = ctx
	}
}
