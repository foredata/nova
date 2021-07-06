package recovery

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/foredata/nova/debug/logs"
)

type RecoverFunc func(ctx context.Context, err error)

var onRecover = func(ctx context.Context, err error) {
	logs.Fatalf("panic err=%+v, stack=%s", err, debug.Stack())
}

// SetDefault 设置默认recover函数
func SetDefault(f RecoverFunc) {
	onRecover = f
}

// Recover 用于保护go routine,默认打印日志
// example:
// go func() {
//		defer recovery.Recover()
// }
func Recover(opts ...Option) {
	o := &Options{ctx: context.Background()}
	for _, fn := range opts {
		fn(o)
	}

	e := recover()
	err, ok := e.(error)
	if !ok {
		err = fmt.Errorf("%+v", e)
	}

	onRecover(o.ctx, err)
	if o.err != nil {
		*o.err = err
	}
}
