package idgen

import (
	"sync"
	"time"
)

// Options 可选参数
type Options struct {
	useSecond bool        // 是否使用秒模式,默认毫秒
	clock     Clocker     //
	lock      sync.Locker // 锁
}

type Option func(o *Options)

// WithSecond 使用秒级时间戳
func WithSecond() Option {
	return func(o *Options) {
		o.useSecond = true
	}
}

// WithClock 自定义clock
func WithClock(c Clocker) Option {
	return func(o *Options) {
		o.clock = c
	}
}

// WithLock 自定义lock
func WithLock(l sync.Locker) Option {
	return func(o *Options) {
		o.lock = l
	}
}

func newOptions(opts ...Option) *Options {
	o := &Options{}
	for _, fn := range opts {
		fn(o)
	}

	if o.clock == nil {
		o.clock = gStdClock
	}

	if o.lock == nil {
		o.lock = &sync.Mutex{}
	}

	return o
}

var gStdClock = &stdClock{}

type stdClock struct {
}

func (stdClock) Now() time.Time {
	return time.Now()
}
