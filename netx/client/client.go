package client

import (
	"context"
	"errors"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/netx/loadbalance"

	// 强制注册codec
	_ "github.com/foredata/nova/netx/codec"
)

// some error
var (
	ErrNoInstances     = errors.New("no instances")
	ErrInvalidCallback = errors.New("invalid callback")
)

// New 创建client
func New(opts ...Option) netx.Client {
	o := newOptions(opts...)
	cli := &client{opts: o}
	return cli
}

type client struct {
	opts  *Options
	cache filterCache
}

func (c *client) Call(ctx context.Context, req netx.Request, opts ...netx.CallOption) (netx.Response, error) {
	o := newCallOptions(opts...)
	if o.DialTimeout == 0 {
		o.DialTimeout = c.opts.Config.GetDialTimeout(ctx, req)
	}
	if o.CallTimeout == 0 {
		o.CallTimeout = c.opts.Config.GetCallTimeout(ctx, req)
	}

	if req.SeqID() == 0 {
		req.SetSeqID(netx.NewSeqID())
	}

	service := req.Service()
	if c.opts.Proxy != "" {
		service = c.opts.Proxy
	}

	picker, err := c.resolve(ctx, service)
	if err != nil {
		return nil, err
	}

	callback, future := toCallback(o.Callback)
	if callback == nil {
		return nil, ErrInvalidCallback
	}

	var retry Retryer
	if o.RetryPolicy != nil {
		retry = newRetry(o.RetryPolicy, c, ctx, req, picker, o)
	}

	if err := c.opts.caller.Register(req, callback, o.CallTimeout, retry); err != nil {
		return nil, err
	}

	var lastErr error
	// 连接失败,默认会自动重试
	for i := 0; i < c.opts.Failover+1; i++ {
		lastErr = c.sendRequest(ctx, picker, req, o)
	}

	if lastErr != nil {
		c.opts.caller.Unregister(req.SeqID())
		if future != nil {
			future.Done(nil, lastErr)
		}
	}

	if future != nil {
		rsp, err := future.Wait()
		future.Recycle()
		return rsp, err
	}

	return nil, lastErr
}

// resolve 解析地址
func (c *client) resolve(ctx context.Context, service string) (loadbalance.Picker, error) {
	entry, err := c.opts.Resolver.Resolve(ctx, service)
	if err != nil {
		return nil, err
	}

	if entry.Empty() {
		return nil, ErrNoInstances
	}

	filter := c.opts.Filter
	if filter != nil {
		if filter.HashCode() == 0 {
			entry, err = filter.Do(entry)
		} else {
			entry, err = c.cache.Get(service, filter, entry)
		}

		if err != nil {
			return nil, err
		}

		if entry.Empty() {
			return nil, ErrNoInstances
		}
	}

	return c.opts.Balancer.Pick(entry)
}

// sendRequest 发送消息
func (c *client) sendRequest(ctx context.Context, picker loadbalance.Picker, req netx.Request, o *CallOptions) error {
	ins, err := picker.Next()
	if err != nil {
		return err
	}

	conn, err := c.opts.ConnPool.Get(ctx, ins, c.opts.Tran, o)
	if err != nil {
		return err
	}

	if err := conn.Send(req); err != nil {
		return err
	}

	return nil
}

func (c *client) Close() error {
	return nil
}
