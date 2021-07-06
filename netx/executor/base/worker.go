package base

import (
	"sync"

	"github.com/foredata/nova/netx"
)

// NewWorker 创建Worker
func NewWorker() *Worker {
	w := &Worker{}
	w.cnd = sync.NewCond(&w.mux)
	return w
}

// Worker 执行协程
type Worker struct {
	queue Queue
	mux   sync.Mutex
	cnd   *sync.Cond
	quit  bool
}

// Close 退出
func (w *Worker) Close() error {
	w.mux.Lock()
	w.quit = true
	w.mux.Unlock()
	w.cnd.Signal()
	return nil
}

// Post 投递一条任务
func (w *Worker) Post(task netx.Runnable) {
	w.mux.Lock()
	w.queue.Push(task)
	w.mux.Unlock()
	w.cnd.Signal()
}

// Run 循环执行任务
func (w *Worker) Run() {
	q := Queue{}
	for {
		w.mux.Lock()

		for !w.quit && w.queue.Empty() {
			w.cnd.Wait()
		}
		quit := w.quit
		w.queue.Swap(&q)

		w.mux.Unlock()

		q.Process()

		if quit {
			break
		}
	}
}

func (w *Worker) RunWithWaitGroup(wg *sync.WaitGroup) {
	wg.Add(1)
	w.Run()
	wg.Done()
}
