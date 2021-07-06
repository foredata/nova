package gopool

import (
	"runtime"
	"sync"
	"sync/atomic"
)

var gPool = NewPool(0)

// Go 执行回调函数
func Go(fn Func) {
	gPool.Post(fn)
}

// Func 回调函数
type Func func()

// Pool goroutine缓存队列,限制goroutine最大数量
// TODO: support cancel task?
type Pool interface {
	Post(fn Func) // 提交任务，等待异步执行
	Wait()        // 等待所有goroutine完成
}

func NewPool(max int) Pool {
	if max <= 0 {
		max = runtime.NumCPU() * 2
	}

	p := &pool{max: int32(max), tasks: newQueue()}
	return p
}

type pool struct {
	tasks *lfQueue       // 任务队列
	wg    sync.WaitGroup // 用于等待所有任务完成
	max   int32          // 最大协程数
	num   int32          // 当前协程数
}

func (p *pool) Post(task Func) {
	if atomic.LoadInt32(&p.num) < p.max {
		// 协程数有可能会突破p.max但影响不大
		atomic.AddInt32(&p.num, 1)
		p.wg.Add(1)
		go p.workLoop(task)
	} else {
		p.tasks.Enqueue(task)
	}
}

func (p *pool) Wait() {
	p.wg.Wait()
}

func (p *pool) workLoop(task Func) {
	task()
	for {
		next := p.tasks.Dequeue()
		if next == nil {
			break
		}
		task = next.(Func)
		task()
	}

	atomic.AddInt32(&p.num, -1)
	p.wg.Done()
}
