package executor

import (
	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/netx/executor/gorunner"
)

func init() {
	SetDefault(gorunner.New())
}

// gDefault 默认全局Executor
var gDefault netx.Executor

// Default 获取默认Executor
func Default() netx.Executor {
	return gDefault
}

// SetDefault 设置默认executor,非线程安全,仅初始化设置
func SetDefault(ex netx.Executor) {
	gDefault = ex
}
