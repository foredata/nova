package hashing

import (
	"sync"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/netx/executor/base"
)

// New 通过hash方式路由到对应worker,要求task实现Index接口
func New(num int) netx.Executor {
	if num < 1 {
		num = 1
	}

	e := &hashingExecutor{}
	for i := 0; i < num; i++ {
		w := base.NewWorker()
		go w.RunWithWaitGroup(&e.wg)
	}

	return e
}

// 需要Task实现Index接口
type indexed interface {
	Index() int
}

type hashingExecutor struct {
	workers []*base.Worker
	wg      sync.WaitGroup
}

func (h *hashingExecutor) Name() string {
	return "hashing"
}

func (h *hashingExecutor) Close() error {
	for _, w := range h.workers {
		w.Close()
	}
	h.wg.Wait()
	return nil
}

func (h *hashingExecutor) Post(task netx.Runnable) error {
	index := 0
	if t, ok := task.(indexed); ok {
		index = t.Index() % len(h.workers)
	}

	h.workers[index].Post(task)
	return nil
}
