package transport

import (
	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/netx/transport/gpc"
)

func init() {
	SetDefault(gpc.New)
}

var gDefault netx.Factory

// SetDefault 设置默认Factory
func SetDefault(fn netx.Factory) {
	gDefault = fn
}

// New 新建Transport
func New() netx.Tran {
	return gDefault()
}
