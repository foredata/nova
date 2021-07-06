package client

import (
	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/netx/discovery"
	"github.com/foredata/nova/netx/executor"
	"github.com/foredata/nova/netx/loadbalance"
	"github.com/foredata/nova/netx/loadbalance/random"
	"github.com/foredata/nova/netx/processor"
	"github.com/foredata/nova/netx/protocol"
	"github.com/foredata/nova/netx/transport"
)

// Options 可选配置信息
type Options struct {
	Tran     netx.Tran            //
	Protocol netx.Protocol        // 默认编码协议
	Exec     netx.Executor        //
	Resolver discovery.Resolver   //
	Balancer loadbalance.Balancer //
	Filter   loadbalance.Filter   // 用于过滤
	Config   Configer             // 配置信息
	Proxy    string               // 代理服务名
	Failover int                  // 故障转移次数
	ConnPool ConnPool             //
	caller   Caller               //
}

type Option func(*Options)

func newOptions(opts ...Option) *Options {
	o := &Options{}
	for _, fn := range opts {
		fn(o)
	}

	if o.caller == nil {
		o.caller = newCaller()
	}

	if o.Protocol == nil {
		o.Protocol = protocol.Default()
	}

	if o.Tran == nil {
		if o.Exec == nil {
			o.Exec = executor.Default()
		}

		filter := processor.NewFilter(o.Exec, o.caller, &detector{o.Protocol})
		tran := transport.New()
		tran.AddFilters(filter)
		o.Tran = tran
	}

	if o.ConnPool == nil {
		o.ConnPool = newConnPool()
	}

	if o.Config == nil {
		o.Config = gDefaultConfig
	}

	if o.Balancer == nil {
		o.Balancer = random.New()
	}

	return o
}

func WithTran(t netx.Tran) Option {
	return func(o *Options) {
		o.Tran = t
	}
}

func WithProtocol(p netx.Protocol) Option {
	return func(o *Options) {
		o.Protocol = p
	}
}

func WithResolver(r discovery.Resolver) Option {
	return func(o *Options) {
		o.Resolver = r
	}
}

func WithBalancer(b loadbalance.Balancer) Option {
	return func(o *Options) {
		o.Balancer = b
	}
}

func WithFilter(v loadbalance.Filter) Option {
	return func(o *Options) {
		o.Filter = v
	}
}

func WithConfig(v Configer) Option {
	return func(o *Options) {
		o.Config = v
	}
}

func WithProxy(p string) Option {
	return func(o *Options) {
		o.Proxy = p
	}
}

func WithConnPool(p ConnPool) Option {
	return func(o *Options) {
		o.ConnPool = p
	}
}

func WithFailover(v int) Option {
	return func(o *Options) {
		o.Failover = v
	}
}
