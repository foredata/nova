package ebus

import (
	"context"
	"reflect"
	"sync"
)

// 全局EventBus
var gBus = bus{}

// Listen 注册事件回调,async标记同步执行还是异步执行
func Listen(ev string, cb Callback, async bool) {
	gBus.Listen(ev, cb, async)
}

// Remove 删除事件回调
func Remove(ev string, cb Callback) {
	gBus.Remove(ev, cb)
}

// Dispatch 触发事件,阻塞调用
func Dispatch(ctx context.Context, ev Event) {
	gBus.Dispatch(ctx, ev)
}

// Event 事件接口
type Event interface {
	Name() string
}

// Callback 事件回调
type Callback func(ctx context.Context, ev Event)

// Bus 事件总线
type Bus interface {
	Listen(ev string, cb Callback, async bool)
	Remove(ev string, cb Callback)
	Dispatch(ctx context.Context, ev Event)
}

type handler struct {
	cb    Callback
	async bool
}

type bus struct {
	mux       sync.RWMutex
	callbacks map[string][]handler
}

func (b *bus) Listen(ev string, cb Callback, async bool) {
	b.mux.Lock()
	b.callbacks[ev] = append(b.callbacks[ev], handler{cb: cb, async: async})
	b.mux.Unlock()
}

func (b *bus) Remove(ev string, cb Callback) {
	b.mux.Lock()
	target := reflect.ValueOf(cb).Pointer()
	callbacks := b.callbacks[ev]
	for i, h := range callbacks {
		if reflect.ValueOf(h.cb).Pointer() == target {
			callbacks = append(callbacks[:i], callbacks[i+1:]...)
			b.callbacks[ev] = callbacks
			break
		}
	}
	b.mux.Unlock()
}

func (b *bus) Dispatch(ctx context.Context, ev Event) {
	// b.mux.RLock()
	// callbacks := b.callbacks[ev.Name()]
	// if len(callbacks) > 0 {
	// 	for _, h := range callbacks {
	// 		if h.async {
	// 			gopool.Go(func() {
	// 				h.cb(ctx, ev)
	// 			})
	// 		} else {
	// 			h.cb(ctx, ev)
	// 		}
	// 	}
	// }

	// b.mux.RUnlock()
}
