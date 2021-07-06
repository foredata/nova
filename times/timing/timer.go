package timing

import (
	"sync"
	"sync/atomic"
	"time"
)

var (
	// gTimerMap 使用map映射一次可以保证多次Stop时并发安全，但会有性能损耗
	gTimerMap = make(map[uint64]Timer)
	gTimerMux sync.RWMutex
	gDefault  Engine = NewEngine(nil)
	gIdMax    uint64
)

// SetDefault 设置默认engine
func SetDefault(e Engine) {
	gDefault = e
}

// NewTimer 根据过期时间创建定时器,如果需要再次重置，需要先关闭原timer，再创建新timer
func NewTimer(expired time.Time, cb Callback, data interface{}, opts ...Option) ID {
	t := newTimer(expired.UnixNano()/int64(time.Millisecond), 0, cb, data, newOptions(opts...))
	if t != nil {
		return t.ID()
	}

	return 0
}

// NewDelayer 创建timer,delay为延迟时间
func NewDelayer(delay time.Duration, cb Callback, data interface{}, opts ...Option) ID {
	o := newOptions(opts...)
	expired := o.engine.Now() + delay.Milliseconds()
	t := newTimer(expired, 0, cb, data, o)
	if t != nil {
		return t.ID()
	}

	return 0
}

// NewTicker 启动定时器,多次执行,本次执行完后会重新注册
func NewTicker(d time.Duration, cb Callback, data interface{}, opts ...Option) ID {
	t := newTimer(0, d.Milliseconds(), cb, data, newOptions(opts...))
	if t != nil {
		return t.ID()
	}

	return 0
}

// Stop 关闭定时器
func Stop(timerID uint64) {
	if timerID == 0 {
		// 忽略无效id
		return
	}
	gTimerMux.Lock()
	timer := gTimerMap[timerID]
	if timer != nil {
		timer.Stop()
	}
	gTimerMux.Unlock()
}

// Callback 回调函数,data 为传入附加参数,如果不需要额外参数可以传nil
type Callback func(data interface{})

// ID 定时器唯一ID,自增且非零
type ID = uint64

// Timer 定时器
//	Timer不能Reset,只能Stop后重新创建新Timer,Timer释放后会放到Pool中复用
//	使用时用户持有ID，而非Timer指针，使用更加安全，但会有一点额外的性能损耗
type Timer interface {
	// ID 全局唯一ID，非零自增
	ID() ID
	// Stop 停止timer or ticker
	Stop()
	// 触发定时器回调,内部调用
	invoke()
}

var gTimerPool = sync.Pool{
	New: func() interface{} {
		return &timer{}
	},
}

func newTimer(expired, interval int64, callback Callback, data interface{}, o *Options) Timer {
	id := atomic.AddUint64(&gIdMax, 1)
	if id == 0 {
		id = atomic.AddUint64(&gIdMax, 1)
	}

	t := gTimerPool.Get().(*timer)
	t.id = id
	t.expired = expired
	t.interval = interval
	t.callback = callback
	t.engine = o.engine
	t.data = data

	if t.engine.Start(t) {
		gTimerMux.Lock()
		gTimerMap[id] = t
		gTimerMux.Unlock()
		return t
	} else {
		t.setStatus(statusIdle)
		gTimerPool.Put(t)
		return nil
	}
}

type status uint8

const (
	statusIdle     status = iota // 空闲状态
	statusTiming                 // 计时中,等待过期
	statusExec                   // 时间过期，等待执行
	statusStopping               // 手动关闭，等待结束,结束后会被回收
)

// Timer 定时器
type timer struct {
	list     *bucket
	prev     *timer
	next     *timer
	engine   Engine      //
	id       ID          // 唯一ID
	expired  int64       // 过期时间
	interval int64       // ticker使用
	callback Callback    // 回调函数
	status   status      // 标识是否被主动关闭
	data     interface{} //
	mux      sync.Mutex  // 调用stop时确保安全
}

func (t *timer) ID() ID {
	return t.id
}

// Stop 线程安全,确保未执行完的timer一定不会被再次使用
// 1: statusTiming状态可直接释放并回收
// 2: statusExec状态不能被释放，只能标记stopping状态，等待执行时忽略执行并回收
func (t *timer) Stop() {
	t.mux.Lock()

	if t.status == statusTiming {
		if t.engine.Stop(t) {
			t.Recyle()
		} else {
			t.setStatus(statusStopping)
		}
	} else if t.status == statusExec {
		t.setStatus(statusStopping)
	}

	t.mux.Unlock()
}

func (t *timer) invoke() {
	t.mux.Lock()

	canRecyle := false
	if t.status == statusExec {
		t.callback(t.data)
		if t.interval != 0 {
			t.engine.Start(t)
		} else {
			canRecyle = true
		}
	} else if t.status == statusStopping {
		canRecyle = true
	}

	if canRecyle {
		gTimerMux.Lock()
		t.Recyle()
		gTimerMux.Unlock()
	}

	t.mux.Unlock()
}

func (t *timer) setStatus(s status) {
	t.status = s
}

func (t *timer) Recyle() {
	t.data = nil
	t.setStatus(statusIdle)
	delete(gTimerMap, t.id)
	gTimerPool.Put(t)
}
