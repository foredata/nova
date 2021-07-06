package single

import (
	"sync"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/netx/executor/base"
)

// New 创建单线程执行器
func New() netx.Executor {
	e := &singleExecutor{}
	e.worker = base.NewWorker()
	e.wg.Add(1)
	go func() {
		e.worker.Run()
		e.wg.Done()
	}()

	return e
}

// 单线程执行
type singleExecutor struct {
	worker *base.Worker
	wg     sync.WaitGroup
}

func (s *singleExecutor) Name() string {
	return "single"
}

func (s *singleExecutor) Close() error {
	s.worker.Close()
	s.wg.Wait()
	return nil
}

func (s *singleExecutor) Post(task netx.Runnable) error {
	s.worker.Post(task)
	return nil
}
