package circuit

import (
	"time"
)

// https://github.com/rubyist/circuitbreaker
// https://github.com/sony/gobreaker
// https://github.com/afex/hystrix-go
//
// https://github.com/Netflix/Hystrix/wiki/
// https://github.com/openbilibili/go-common
// https://zhuanlan.zhihu.com/p/58428026
// https://blog.csdn.net/tongtong_use/article/details/78611225
// https://segmentfault.com/a/1190000005988895
// https://techblog.constantcontact.com/software-development/circuit-breakers-and-microservices/

// https://yangxikun.com/golang/2019/08/10/golang-circuit.html
// https://github.com/cep21/circuit
// https://github.com/mercari/go-circuitbreaker
type Breaker interface {
	Allow() bool
	Succeed()
	Fail()
	Timeout()
	State() State
	Counter() Counter
	Reset()
}

// Panel breaker集合
type Panel interface {
	Remove(key string)
	Allow(key string) bool
	Succeed(key string)
	Fail(key string)
	Timeout(key string)
	Close() error
}

// Counts 统计信息
type Counts struct {
	Successes   int64         // 成功数
	Failures    int64         // 失败数
	Timeouts    int64         // 超时数
	ConseErrors int64         // 连续错误个数,consecutive
	ConseTime   time.Duration // 连续错误时间
}

func (c *Counts) Errors() int64 {
	return c.Failures + c.Timeouts
}

func (c *Counts) Samples() int64 {
	return c.Successes + c.Failures + c.Timeouts
}

func (c *Counts) ErrorRate() float64 {
	errors := c.Errors()
	samples := errors + c.Successes
	if samples == 0 {
		return 0
	}

	return float64(errors) / float64(samples)
}

func (c *Counts) reset() {
	c.Successes = 0
	c.Failures = 0
	c.Timeouts = 0
	c.ConseErrors = 0
	c.ConseTime = 0
}

// Counter 用于breaker统计error,success数据
type Counter interface {
	Counts() Counts
	Succeed()
	Fail()
	Timeout()
	Reset()
	Tick()
}

// TripFunc is a function called by a breaker when error appear and
// determines whether the breaker should trip.
type TripFunc func(Counter) bool

// StateChangedHandler use to notify state change
type StateChangedHandler func(oldState, newState State, c Counter)

// NowFunc used to get now time
type NowFunc func() time.Time

var gNowFunc = time.Now

// Now get now time
func Now() time.Time {
	return gNowFunc()
}

// NowUnixNano 返回当前nano时间戳
func NowUnixNano() int64 {
	return Now().UnixNano()
}
