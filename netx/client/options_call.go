package client

import (
	"time"

	"github.com/foredata/nova/netx"
)

type CallOptions = netx.CallOptions
type CallOption = netx.CallOption

func newCallOptions(opts ...CallOption) *CallOptions {
	o := &CallOptions{}
	for _, fn := range opts {
		fn(o)
	}
	return o
}

// WithDialTimeout .
func WithDialTimeout(t time.Duration) CallOption {
	return func(co *CallOptions) {
		co.DialTimeout = t
	}
}

// WithCallTimeout .
func WithCallTimeout(t time.Duration) CallOption {
	return func(co *CallOptions) {
		co.CallTimeout = t
	}
}

// WithCallback .
func WithCallback(v interface{}) CallOption {
	return func(co *CallOptions) {
		co.Callback = v
	}
}

// WithRetryPolicy .
func WithRetryPolicy(p netx.RetryPolicy) CallOption {
	return func(co *CallOptions) {
		co.RetryPolicy = p
	}
}
