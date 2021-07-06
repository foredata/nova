package loadbalance

import (
	"github.com/foredata/nova/netx/discovery"
)

// Picker picks an instance for next RPC call.
type Picker interface {
	Next() (discovery.Instance, error)
	Recycle()
}

// Filter 过滤instance
// 	HashCode用于cache,0则不cache
type Filter interface {
	HashCode() uint32
	Do(*discovery.Result) (*discovery.Result, error)
}

// Balancer generates pickers for the given service discovery result.
type Balancer interface {
	Pick(*discovery.Result) (Picker, error)
}
