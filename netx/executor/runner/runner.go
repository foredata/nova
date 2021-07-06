package runner

import "github.com/foredata/nova/netx"

// New ...
func New() netx.Executor {
	return &runnerExecutor{}
}

// runnerExecutor 直接运行task
type runnerExecutor struct {
}

func (r *runnerExecutor) Name() string {
	return "runner"
}

func (r *runnerExecutor) Close() error {
	return nil
}

func (e *runnerExecutor) Post(task netx.Runnable) error {
	return task.Run()
}
