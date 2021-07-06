package server

import (
	"os"
	"time"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/netx/executor"
	"github.com/foredata/nova/netx/processor"
	"github.com/foredata/nova/netx/registry"
	"github.com/foredata/nova/netx/transport"
	"github.com/foredata/nova/pkg/xid"
)

const (
	defaultRegistryTTL = time.Second * 15
)

// Options 可选配置参数
type Options struct {
	ID          string            // 唯一ID,如果不指定,则随机生成
	Name        string            // 服务名
	Version     string            // 服务版本
	Metadata    map[string]string // Meta
	Addr        string            // 监听地址
	Tran        netx.Tran         // Transport
	Detector    netx.Detector     // 协议探测,默认自动探测
	Router      netx.Router       // 路由
	Exec        netx.Executor     // 调度器,默认每条消息一个go routine并发执行
	Node        *registry.Node    // node配置信息
	Registry    registry.Registry // 服务注册
	RegistryTTL time.Duration     // 注册过期时间
	Signals     []os.Signal       // 需要监听的事件,默认syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT
	Modules     []netx.Module     // 扩展模块
	Binder      Binder            //
	Validator   Validator         //
	Codec       netx.CodecType    // 默认编解码协议
}

type Option func(o *Options)

func newOptions(opts ...Option) *Options {
	o := &Options{
		Codec: netx.CodecTypeJson,
	}
	for _, fn := range opts {
		fn(o)
	}

	if o.ID == "" {
		o.ID = xid.New().String()
	}

	if o.Router == nil {
		o.Router = NewRouter()
	}

	if o.Detector == nil {
		o.Detector = newDetector()
	}

	if o.Tran == nil {
		if o.Exec == nil {
			o.Exec = executor.Default()
		}

		filter := processor.NewFilter(o.Exec, o.Router, o.Detector)
		tran := transport.New()
		tran.AddFilters(filter)
		o.Tran = tran
	}

	if o.Binder == nil {
		o.Binder = defaultBinder
	}

	if o.RegistryTTL == 0 {
		o.RegistryTTL = defaultRegistryTTL
	}

	return o
}

func WithID(id string) Option {
	return func(o *Options) {
		o.ID = id
	}
}

func WithName(name string) Option {
	return func(o *Options) {
		o.Name = name
	}
}

func WithVersion(v string) Option {
	return func(o *Options) {
		o.Version = v
	}
}

func WithMetadata(v map[string]string) Option {
	return func(o *Options) {
		o.Metadata = v
	}
}

func WithAddr(addr string) Option {
	return func(o *Options) {
		o.Addr = addr
	}
}

func WithTran(t netx.Tran) Option {
	return func(o *Options) {
		o.Tran = t
	}
}

func WithDetector(d netx.Detector) Option {
	return func(o *Options) {
		o.Detector = d
	}
}

func WithRouter(r netx.Router) Option {
	return func(o *Options) {
		o.Router = r
	}
}

func WithExec(e netx.Executor) Option {
	return func(o *Options) {
		o.Exec = e
	}
}

func WithNode(n *registry.Node) Option {
	return func(o *Options) {
		o.Node = n
	}
}

func WithRegistry(r registry.Registry) Option {
	return func(o *Options) {
		o.Registry = r
	}
}

func WithSignals(s ...os.Signal) Option {
	return func(o *Options) {
		o.Signals = s
	}
}

func WithModules(m ...netx.Module) Option {
	return func(o *Options) {
		o.Modules = m
	}
}

func WithBinder(b Binder) Option {
	return func(o *Options) {
		o.Binder = b
	}
}

func WithValidator(v Validator) Option {
	return func(o *Options) {
		o.Validator = v
	}
}

func WithCodec(v netx.CodecType) Option {
	return func(o *Options) {
		o.Codec = v
	}
}
