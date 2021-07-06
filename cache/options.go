package cache

import (
	"context"
)

// GetFunc 当缓存不存在时,加载回调函数,可使用singleflight方式避免缓存击穿
type LoadFunc func(ctx context.Context, key string) ([]byte, error)

type GetOptions struct {
	Load LoadFunc
}

type GetOption func(o *GetOptions)

func newGetOptions(opts ...GetOption) *GetOptions {
	o := &GetOptions{}
	for _, fn := range opts {
		fn(o)
	}

	return o
}

// WithLoad .
func WithLoad(fn LoadFunc) GetOption {
	return func(o *GetOptions) {
		o.Load = fn
	}
}
