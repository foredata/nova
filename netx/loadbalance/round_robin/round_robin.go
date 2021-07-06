package roundrobin

import (
	"math/rand"
	"sync"
	"sync/atomic"

	"github.com/foredata/nova/netx/discovery"
	"github.com/foredata/nova/netx/loadbalance"
)

var gBalancer = &balancer{}

var gPickerPool = sync.Pool{
	New: func() interface{} {
		return &picker{}
	},
}

type picker struct {
	result *discovery.Result
	index  int32
}

func (p *picker) Reset(result *discovery.Result) {
	p.result = result
	p.index = rand.Int31()
}

func (p *picker) Next() (discovery.Instance, error) {
	idx := atomic.AddInt32(&p.index, 1)
	return p.result.Instances()[idx%int32(p.result.Len())], nil
}

func (p *picker) Recycle() {
	gPickerPool.Put(p)
}

type balancer struct {
}

func (b *balancer) Pick(result *discovery.Result) (loadbalance.Picker, error) {
	p := gPickerPool.Get().(*picker)
	p.Reset(result)
	return p, nil
}

func New() loadbalance.Balancer {
	return gBalancer
}
