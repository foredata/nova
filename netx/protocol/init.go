package protocol

import (
	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/netx/protocol/http1"
	"github.com/foredata/nova/netx/protocol/rpc"
	"github.com/foredata/nova/pkg/bytex"
)

func init() {
	// 注册已知协议
	Register(rpc.New())
	Register(http1.New())

	SetDefault(http1.New())
}

var gDefault netx.Protocol
var gList []netx.Protocol
var gDict = make(map[string]netx.Protocol)

// Register 注册协议,同名则会忽略,非线程安全,仅可以在启动时初始化
func Register(p netx.Protocol) {
	if _, ok := gDict[p.Name()]; ok {
		return
	}
	gDict[p.Name()] = p
	gList = append(gList, p)
}

// Get 根据名称,返回指定协议,区分大小写
func Get(name string) netx.Protocol {
	return gDict[name]
}

// Detect 自动探测协议
func Detect(r bytex.Peeker) netx.Protocol {
	for _, p := range gList {
		if ok := p.Detect(r); ok {
			return p
		}
	}

	return nil
}

// Default 全局默认Protocol
func Default() netx.Protocol {
	return gDefault
}

// SetDefault 设置默认
func SetDefault(p netx.Protocol) {
	gDefault = p
}
