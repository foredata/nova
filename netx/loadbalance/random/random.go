package random

import (
	"math/rand"
	"sync"

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
}

func (p *picker) Reset(result *discovery.Result) {
	p.result = result
}

func (p *picker) Next() (discovery.Instance, error) {
	i := rand.Intn(len(p.result.Instances()))
	return p.result.Instances()[i], nil
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
