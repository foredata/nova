package timing

import (
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/foredata/nova/times/clock"
)

// Engine 毫秒级定时器管理
type Engine interface {
	// 当前时间戳
	Now() int64
	// 启动定时器
	Start(timer Timer) bool
	// 关闭定时器
	Stop(timer Timer) bool
	// 关闭Engine
	Close() error
}

type Config struct {
	Precision int64       // 执行精度，单位毫秒，默认为1
	Exec      Executor    // 执行器，默认单独协程中执行过期timer
	Clock     clock.Clock // 用于获取当前时间，默认clock.New()
	WheelNum  int         // 默认wheel个数,默认为3
}

// NewEngine 创建新Engine
func NewEngine(conf *Config) Engine {
	if conf == nil {
		conf = &Config{}
	}
	if conf.Precision <= 0 {
		conf.Precision = 1
	}
	if conf.WheelNum <= 0 {
		conf.WheelNum = 3
	}
	if conf.Exec == nil {
		conf.Exec = NewExecutor()
	}
	if conf.Clock == nil {
		conf.Clock = clock.New()
	}

	eng := &engine{
		exit:      0,
		precision: conf.Precision,
		clock:     conf.Clock,
		exec:      conf.Exec,
		timestamp: toMillsecond(conf.Clock.Now()),
	}

	for i := 0; i < conf.WheelNum; i++ {
		eng.newWheel()
	}

	eng.Run()
	return eng
}

// Hierarchical Time Wheel,毫秒精度
type engine struct {
	wheels    []*wheel    // 时间轮
	precision int64       // 精度,单位毫秒
	timestamp int64       // 当前时间戳,单位毫秒
	maximum   int64       // 当前能保存的时间戳范围
	count     int         // timer个数
	exec      Executor    // 执行器
	clock     clock.Clock // 用于获取当前时间
	exit      int32       // engine是否退出
	mux       sync.Mutex  //
}

func (e *engine) Now() int64 {
	now := toMillsecond(e.clock.Now())
	return now
}

func (e *engine) Start(t Timer) (result bool) {
	tt, ok := t.(*timer)
	if !ok {
		return false
	}

	e.mux.Lock()
	if tt.interval != 0 {
		// ticker
		tt.expired = e.timestamp + tt.interval
	}
	if tt.expired > e.timestamp {
		tt.setStatus(statusTiming)
		e.push(tt)
		e.count++
		result = true
	} else {
		result = false
	}
	e.mux.Unlock()

	return
}

func (e *engine) Stop(t Timer) (result bool) {
	tt, ok := t.(*timer)
	if !ok {
		return false
	}
	e.mux.Lock()
	if tt.list != nil {
		result = true
		tt.list.Remove(tt)
		tt.list = nil
		e.count--
		result = true
	} else {
		result = false
	}

	e.mux.Unlock()
	return
}

func (e *engine) Close() error {
	atomic.StoreInt32(&e.exit, 1)
	return nil
}

func (e *engine) Run() {
	go e.loop()
}

func (e *engine) loop() {
	for atomic.LoadInt32(&e.exit) == 0 {
		e.mux.Lock()
		now := toMillsecond(e.clock.Now())
		if now < e.timestamp {
			// 时间发生回滚,Admin强制调整时间导致?
			e.rebuild(now)
		} else {
			e.tick(now)
		}
		e.mux.Unlock()
		//
		sleep := now - e.timestamp
		if sleep == 0 {
			sleep = e.precision
		}
		e.clock.Sleep(time.Duration(sleep * int64(time.Millisecond)))
	}
}

func (e *engine) process(pendings *bucket) {
	if pendings.Len() == 0 {
		return
	}

	e.count -= pendings.Len()
	timers := make([]Timer, 0, pendings.Len())
	for iter := pendings.Front(); iter != nil; {
		t := iter
		iter = iter.next
		pendings.Remove(t)
		t.setStatus(statusExec)
		timers = append(timers, t)
	}
	e.exec.Post(timers)
}

// 重新构建
func (e *engine) rebuild(now int64) {
	var pendings bucket

	e.timestamp = now
	// 发生时间回调,重新构建所有时间,需要重新计算所有timer
	for _, wheel := range e.wheels {
		wheel.index = 0
		for _, slot := range wheel.slots {
			pendings.merge(slot)
		}
	}

	// 重新计算时间
	for iter := pendings.Front(); iter != nil; {
		timer := iter
		iter = iter.next
		if timer.expired >= now {
			pendings.Remove(timer)
			e.push(timer)
		}
	}
	e.process(&pendings)
}

func (e *engine) tick(now int64) {
	var pendings bucket

	ticks := (now - e.timestamp) / e.precision
	if ticks <= 0 {
		return
	}
	e.timestamp += e.precision * ticks

	for i := int64(0); i < ticks; i++ {
		pendings.merge(e.wheels[0].Current())
		e.cascade(&pendings)
	}

	e.process(&pendings)
}

// cascade 前进1个tick,并返回过期的timer,存于pendings中
func (e *engine) cascade(pendings *bucket) {
	for i := 0; i < len(e.wheels); i++ {
		if !e.wheels[i].Step() {
			break
		}

		if i+1 == len(e.wheels) {
			// 溢出,创建新的wheel
			e.newWheel()
			break
		}

		// rehash next wheel
		slots := e.wheels[i+1].Current()
		for iter := slots.Front(); iter != nil; {
			timer := iter
			iter = timer.next
			slots.Remove(timer)

			if timer.expired <= e.timestamp {
				pendings.Push(timer)
			} else {
				e.push(timer)
			}
		}
	}
}

func (e *engine) push(t *timer) {
	var delta int64
	if e.precision != 1 {
		delta = int64(math.Ceil(float64(t.expired-e.timestamp) / float64(e.precision)))
	} else {
		delta = t.expired - e.timestamp
	}

	// 溢出则动态添加wheel
	for delta > e.maximum {
		e.newWheel()
	}

	for i := 0; i < len(e.wheels); i++ {
		wheel := e.wheels[i]
		if delta < wheel.maximum {
			wheel.Push(t, delta)
			break
		}
	}
}

func (e *engine) newWheel() {
	wheel := newWheel(len(e.wheels))
	e.maximum = wheel.maximum
	e.wheels = append(e.wheels, wheel)
}

func toMillsecond(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}
