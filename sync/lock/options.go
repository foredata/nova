package lock

import (
	"time"
)

const (
	DefaultTTL = time.Minute
	TimeoutMax = -1
)

type Options struct {
	TTL     time.Duration // 表示宕机后最长TTL时间后可以重新获得锁,<=0则使用默认值
	Timeout time.Duration // Lock等待超时时间,0则不等待立即返回,TimeoutMax则表示永不超时,直到获取到锁
}

type Option func(o *Options)

func NewOptions(opts ...Option) *Options {
	o := &Options{}
	for _, fn := range opts {
		fn(o)
	}

	return o
}

// TTL set the lock ttl
func TTL(t time.Duration) Option {
	return func(o *Options) {
		o.TTL = t
	}
}

// Timeout set the lock wait timeout
func Timeout(t time.Duration) Option {
	return func(o *Options) {
		o.Timeout = t
	}
}

// wait until get lock
func Blocking() Option {
	return func(o *Options) {
		o.Timeout = TimeoutMax
	}
}
