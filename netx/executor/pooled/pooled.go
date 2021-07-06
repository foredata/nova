package pooled

import (
	"sync"
	"sync/atomic"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/netx/executor/base"
)

// New 协程池
func New(max int) netx.Executor {
	if max < 1 {
		max = 1
	}
	p := &pooledExecutor{max: int32(max)}
	return p
}

// pooledExecutor
//	1:多协程并发执行,无序
//	2:限制最大协程数,超出后放入队列中等待继续执行
// TODO:Lock-Free
// https://www.cnblogs.com/gaochundong/p/lock_free_programming.html
// https://github.com/golang-design/lockfree
type pooledExecutor struct {
	tasks base.Queue     // 待执行队列
	mux   sync.Mutex     // 用于保护tasks
	wg    sync.WaitGroup // 用于等待所有任务完成
	max   int32          // 最大协程数
	num   int32          // 当前协程数
}

func (p *pooledExecutor) Name() string {
	return "pooled"
}

func (p *pooledExecutor) Close() error {
	p.wg.Wait()
	return nil
}

func (p *pooledExecutor) Post(task netx.Runnable) error {
	if atomic.LoadInt32(&p.num) < p.max {
		// 协程数有可能会突破p.max但影响不大
		atomic.AddInt32(&p.num, 1)
		p.wg.Add(1)
		go p.loop(task)
		return nil
	}

	p.mux.Lock()
	p.tasks.Push(task)
	p.mux.Unlock()

	return nil
}

func (p *pooledExecutor) loop(task netx.Runnable) {
	_ = task.Run()

	for {
		p.mux.Lock()
		task = p.tasks.Pop()
		p.mux.Unlock()
		if task == nil {
			break
		}
		_ = task.Run()
	}

	atomic.AddInt32(&p.num, -1)
	p.wg.Done()
}
