package netx

import (
	"context"
	"fmt"
	"math"
	"sync"
)

var gFilterCtxPool = sync.Pool{
	New: func() interface{} {
		return &filterCtx{
			attrs: NewAttributeMap(),
		}
	},
}

// newFilterCtx 创建FilterCtx
func newFilterCtx(filters []Filter, conn Conn, forward bool, cb callback) *filterCtx {
	ctx := gFilterCtxPool.Get().(*filterCtx)
	ctx.init(filters, conn, forward, cb)
	return ctx
}

// callback 用于Next等调用中，能执行对应的HandleRead，HandleWrite函数
type callback func(filter Filter, ctx FilterCtx) error

// filterCtx implement FilterCtx interface{}
type filterCtx struct {
	context.Context
	attrs   AttributeMap
	filters []Filter
	conn    Conn
	data    interface{}
	err     error
	index   int
	forward bool
	cb      callback
}

func (ctx *filterCtx) init(filters []Filter, conn Conn, forward bool, cb callback) {
	ctx.attrs.Clear()
	ctx.filters = filters
	ctx.conn = conn
	ctx.data = nil
	ctx.forward = forward
	ctx.cb = cb
	// 调用next时会自动加一或减一
	if forward {
		ctx.index = -1
	} else {
		ctx.index = len(filters)
	}
}

func (ctx *filterCtx) Attributes() AttributeMap {
	return ctx.attrs
}

func (ctx *filterCtx) Conn() Conn {
	return ctx.conn
}

func (ctx *filterCtx) Data() interface{} {
	return ctx.data
}

func (ctx *filterCtx) SetData(data interface{}) {
	ctx.data = data
}

func (ctx *filterCtx) Error() error {
	return ctx.err
}

func (ctx *filterCtx) SetError(err error) {
	ctx.err = err
}

func (ctx *filterCtx) IsAbort() bool {
	return ctx.index >= math.MaxInt32
}

func (ctx *filterCtx) Abort() {
	if ctx.forward {
		ctx.index = math.MaxInt32
	} else {
		ctx.index = -1
	}
}

// 调用callback
func (ctx *filterCtx) call(index int) error {
	ctx.index = index
	// 这样写可以保证即使没有调用Next也能正确执行到最后一个Filter
	// 如果需要终止，需要主动调用Abort
	//
	// Open,Close,Read是forward
	if ctx.forward {
		for idx := len(ctx.filters); ctx.index < idx; ctx.index++ {
			if err := ctx.cb(ctx.filters[ctx.index], ctx); err != nil {
				ctx.Abort()
				return err
			}
		}
	} else {
		for ; ctx.index >= 0; ctx.index-- {
			if err := ctx.cb(ctx.filters[ctx.index], ctx); err != nil {
				ctx.Abort()
				return err
			}
		}
	}

	return nil
}

func (ctx *filterCtx) Next() error {
	if ctx.forward {
		return ctx.call(ctx.index + 1)
	}

	return ctx.call(ctx.index - 1)
}

func (ctx *filterCtx) Jump(index int) error {
	if index < 0 || index >= len(ctx.filters) {
		return ErrFilterIndexOverflow
	}

	return ctx.call(index)
}

func (ctx *filterCtx) JumpBy(name string) error {
	index := -1
	for idx, filter := range ctx.filters {
		if filter.Name() == name {
			index = idx
			break
		}
	}
	if index == -1 {
		return fmt.Errorf("not found filter,%+v", name)
	}

	return ctx.Jump(index)
}

// Clone 克隆一份,用于转移到其他线程继续执行
func (ctx *filterCtx) Clone() FilterCtx {
	nctx := newFilterCtx(ctx.filters, ctx.conn, ctx.forward, ctx.cb)
	nctx.data = ctx.data
	nctx.err = ctx.err
	nctx.index = ctx.index
	return nctx
}

func (ctx *filterCtx) Call() error {
	err := ctx.Next()
	gFilterCtxPool.Put(ctx)
	return err
}

func (ctx *filterCtx) Recycle() {
	gFilterCtxPool.Put(ctx)
}
