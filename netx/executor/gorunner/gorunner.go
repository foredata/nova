package gorunner

import "github.com/foredata/nova/netx"

// New ...
func New() netx.Executor {
	return &gorunnerExecutor{}
}

// gorunnerExecutor 新起go routine执行task
type gorunnerExecutor struct {
}

func (r *gorunnerExecutor) Name() string {
	return "gorunner"
}

func (r *gorunnerExecutor) Close() error {
	return nil
}

func (e *gorunnerExecutor) Post(task netx.Runnable) error {
	go func() {
		_ = task.Run()
	}()
	return nil
}
