package timing

import (
	"container/list"
	"sync"
)

// Executor timer执行器,默认单独线程中执行
type Executor interface {
	Post(t []Timer)
	Close() error
}

// NewExecutor 默认调度器
func NewExecutor() Executor {
	e := &executor{}
	e.back = list.New()
	e.front = list.New()
	e.cond = sync.NewCond(&e.mux)
	go e.Run()
	return e
}

type element struct {
	timers []Timer
}

type executor struct {
	front *list.List
	back  *list.List
	mux   sync.Mutex
	cond  *sync.Cond
	exit  bool
}

func (e *executor) Post(timers []Timer) {
	e.mux.Lock()
	e.back.PushBack(&element{timers: timers})
	e.mux.Unlock()
	e.cond.Signal()
}

func (e *executor) Close() error {
	e.mux.Lock()
	e.exit = true
	e.mux.Unlock()
	e.cond.Signal()
	return nil
}

func (e *executor) Run() {
	for {
		e.mux.Lock()
		for e.back.Len() == 0 && !e.exit {
			e.cond.Wait()
		}
		e.back, e.front = e.front, e.back
		exit := e.exit
		e.mux.Unlock()

		for e.front.Len() > 0 {
			elem := e.front.Remove(e.front.Front()).(*element)
			for _, t := range elem.timers {
				t.invoke()
			}
		}

		if exit {
			break
		}
	}
}
