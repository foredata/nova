package netx

import (
	"net"
	"time"
)

// NewOptions ...
func NewOptions(opts ...Option) *Options {
	o := &Options{
		Network: "tcp",
		Listen:  stdListen,
		Dial:    stdDial,
	}
	for _, fn := range opts {
		fn(o)
	}

	return o
}

// Options 可选参数
type Options struct {
	Tag             string                 // 额外标签
	Network         string                 // 默认TCP,某些场景下会使用unix socket
	DialTimeout     time.Duration          // 连接超时设置
	DialCallback    DialCallback           // 连接回调
	DialNonBlocking bool                   // 连接是否阻塞,默认阻塞
	Listen          ListenFunc             // Listen
	Dial            DialFunc               // Dial
	Extra           map[string]interface{} // 其他扩展配置
}

func (o *Options) GetExtra(key string) interface{} {
	return o.Extra[key]
}

func (o *Options) GetExtraString(key string, def string) string {
	res, ok := o.Extra[key].(string)
	if ok {
		return res
	}

	return def
}

func (o *Options) GetExtraInt(key string, def int) int {
	res, ok := o.Extra[key].(int)
	if ok {
		return res
	}

	return def
}

// Option ...
type Option func(o *Options)

// ListenFunc 标准的Listen接口
type ListenFunc func(host string, opts *Options) (net.Listener, error)

// DialFunc 标准Dial接口
type DialFunc func(addr string, opts *Options) (net.Conn, error)

// DialCallback Dial回调函数
type DialCallback func(Conn, error)

// stdListen 官方标准listen
func stdListen(host string, opts *Options) (net.Listener, error) {
	return net.Listen(opts.Network, host)
}

// stdDial 官方标准Dial
func stdDial(addr string, opts *Options) (net.Conn, error) {
	return net.DialTimeout(opts.Network, addr, opts.DialTimeout)
}

// WithTag .
func WithTag(tag string) Option {
	return func(o *Options) {
		o.Tag = tag
	}
}

// WithNetwork .
func WithNetwork(v string) Option {
	return func(o *Options) {
		o.Network = v
	}
}

// WithDialTimeout dial超时设置
func WithDialTimeout(v time.Duration) Option {
	return func(o *Options) {
		o.DialTimeout = v
	}
}

// WithDialCallback 设置连接回调
func WithDialCallback(fn DialCallback) Option {
	return func(o *Options) {
		o.DialCallback = fn
	}
}

// WithDialNonBlocking 设置连接非阻塞
func WithDialNonBlocking() Option {
	return func(o *Options) {
		o.DialNonBlocking = true
	}
}

// WithListen 设置Listen函数,默认使用标准tcp
func WithListen(fn ListenFunc) Option {
	return func(o *Options) {
		o.Listen = fn
	}
}

// WithDial 设置Dial函数,默认使用tcp
func WithDial(fn DialFunc) Option {
	return func(o *Options) {
		o.Dial = fn
	}
}

// WithExtra 扩展配置
func WithExtra(key string, value interface{}) Option {
	return func(o *Options) {
		if o.Extra == nil {
			o.Extra = make(map[string]interface{})
		}
		o.Extra[key] = value
	}
}
