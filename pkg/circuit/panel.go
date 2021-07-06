package circuit

import (
	"sync"
	"time"
)

// NewPanel ...
func NewPanel(opts *Options) Panel {
	opts.check()
	p := &panel{
		breakers: make(map[string]Breaker),
		opts:     opts,
		ticker:   time.NewTicker(opts.BucketTime),
	}

	go p.start()
	return p
}

type panel struct {
	breakers map[string]Breaker
	mux      sync.RWMutex
	opts     *Options
	ticker   *time.Ticker
}

func (p *panel) Remove(key string) {
	p.mux.Lock()
	delete(p.breakers, key)
	p.mux.Unlock()
}

func (p *panel) Allow(key string) bool {
	return p.get(key).Allow()
}

func (p *panel) Succeed(key string) {
	p.get(key).Succeed()
}

func (p *panel) Fail(key string) {
	b := p.get(key)
	b.Fail()
}

func (p *panel) Timeout(key string) {
	p.get(key).Timeout()
}

func (p *panel) Close() error {
	p.ticker.Stop()
	return nil
}

func (p *panel) get(key string) Breaker {
	p.mux.RLock()
	if b, ok := p.breakers[key]; ok {
		p.mux.RUnlock()
		return b
	}

	p.mux.Lock()
	b := NewBreaker(p.opts)
	p.breakers[key] = b
	p.mux.Unlock()

	return b
}

func (p *panel) start() {
	for range p.ticker.C {
		p.mux.Lock()
		for _, v := range p.breakers {
			v.Counter().Tick()
		}
		p.mux.Unlock()
	}
}
