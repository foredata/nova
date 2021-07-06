package netx

// Executor 用于执行任务
//	常见线程模型有
//	1:同步执行,runner.New()
//	2:异步单线程,有序执行,single.New()
//	3:异步线程池,无序但线程数不超过最大值,pooled.New()
//	4:异步hash线程,不同功能指定不同线程,hashing.New()
//	5:异步go routine,gorunner.New
type Executor interface {
	Name() string
	Close() error
	Post(task Runnable) error
}

// Runnable 异步任务
type Runnable interface {
	Run() error
}

// RunFunc 实现Task接口,外部可直接把函数转换成RunFunc
type RunFunc func() error

func (fn RunFunc) Run() error {
	return fn()
}
