package client

import (
	"context"
	"sync"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/netx/loadbalance"
)

// RetryPolicy .
type RetryPolicy = netx.RetryPolicy

// Retryer 重试回调
type Retryer interface {
	Allow() bool
	Do(req netx.Request) error
	Recyle()
}

func newRetry(policy RetryPolicy, cli *client, ctx context.Context, req netx.Request, picker loadbalance.Picker, opts *CallOptions) Retryer {
	r := gRetryPool.Get().(*retryer)
	r.policy = policy
	r.cli = cli
	r.ctx = ctx
	r.req = req
	r.picker = picker
	r.opts = opts
	return r
}

var gRetryPool = sync.Pool{
	New: func() interface{} {
		return &retryer{}
	},
}

type retryer struct {
	policy RetryPolicy
	cli    *client
	ctx    context.Context
	req    netx.Request
	picker loadbalance.Picker
	opts   *CallOptions
	count  int
}

func (r *retryer) Allow() bool {
	res := r.policy.Allow(r.ctx, r.req, r.count)
	r.count++
	return res
}

func (r *retryer) Do(req netx.Request) error {
	var lastErr error
	for {
		lastErr = r.cli.sendRequest(r.ctx, r.picker, req, r.opts)
		if lastErr == nil {
			return nil
		}

		if !r.Allow() {
			return lastErr
		}
	}
}

func (r *retryer) Recyle() {
	gRetryPool.Put(r)
}
