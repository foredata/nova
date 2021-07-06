package circuit

import (
	"sync"
	"sync/atomic"
	"time"
)

// NewBreaker ...
func NewBreaker(opts *Options) Breaker {
	return &breaker{opts: opts}
}

type breaker struct {
	opts            *Options
	mux             sync.RWMutex
	counter         Counter
	state           int32
	lastOpenTime    time.Time // the time when the breaker become Open recently
	lastRetryTime   time.Time // last retry time when in HalfOpen State
	halfopenSuccess int32     // consecutive successes when HalfOpen
}

func (b *breaker) Allow() bool {
	switch b.State() {
	case Open:
		now := Now()
		b.mux.RLock()
		if b.lastOpenTime.Add(b.opts.CoolingTime).After(now) {
			b.mux.RUnlock()
			return false
		}

		b.mux.RUnlock()
		b.mux.Lock()
		if b.State() == Open {
			b.setState(Open, HalfOpen)
			b.lastRetryTime = now
		}
		b.mux.Unlock()

	case HalfOpen:
		now := Now()
		b.mux.RLock()
		if b.lastRetryTime.Add(b.opts.DetectTime).After(now) {
			b.mux.RUnlock()
			return false
		}
		b.mux.RUnlock()
		b.mux.Lock()
		if b.State() == HalfOpen {
			b.lastRetryTime = now
		}
		b.mux.Unlock()
	case Closed:
	}
	return true
}

func (b *breaker) Succeed() {
	switch b.State() {
	case Open:
	case HalfOpen:
		b.mux.Lock()
		if b.State() == HalfOpen {
			atomic.AddInt32(&b.halfopenSuccess, 1)
			if atomic.LoadInt32(&b.halfopenSuccess) >= defaultHalfOpenSuccesses {
				b.setState(HalfOpen, Closed)
				b.counter.Reset()
			}
		}
		b.mux.Unlock()
	case Closed:
		b.mux.Lock()
		b.counter.Succeed()
		b.mux.Unlock()
	}
}

func (b *breaker) Fail() {
	b.markError(false)
}

func (b *breaker) Timeout() {
	b.markError(true)
}

func (b *breaker) markError(isTimeout bool) {
	switch b.State() {
	case Open:
	case HalfOpen:
		b.mux.Lock()
		if b.State() == HalfOpen {
			b.setState(HalfOpen, Open)
			b.lastOpenTime = Now()
		}
		b.mux.Unlock()
	case Closed:
		if b.opts.Trip(b.counter) {
			b.mux.Lock()
			if b.State() == Closed {
				b.setState(Closed, Open)
				b.lastOpenTime = Now()
			}
			b.mux.Unlock()
		}
	}
}

func (b *breaker) State() State {
	return State(atomic.LoadInt32(&b.state))
}

func (b *breaker) Counter() Counter {
	return b.counter
}

func (b *breaker) Reset() {
	b.mux.Lock()
	b.counter.Reset()
	atomic.StoreInt32(&b.state, int32(Closed))
	b.mux.Unlock()
}

func (b *breaker) setState(oldState, newState State) {
	if b.opts.OnStateChanged != nil {
		b.opts.OnStateChanged(oldState, newState, b.counter)
	}
	atomic.StoreInt32(&b.state, int32(newState))
}
