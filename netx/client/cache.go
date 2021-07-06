package client

import (
	"fmt"
	"sync"

	"github.com/foredata/nova/netx/discovery"
	"github.com/foredata/nova/netx/loadbalance"
)

type cacheEntry struct {
	entry    *discovery.Result
	hashCode uint32
}

// filterCache 相同的查询条件返回cache结果
type filterCache struct {
	items sync.Map
}

func (f *filterCache) Get(service string, filter loadbalance.Filter, entry *discovery.Result) (*discovery.Result, error) {
	key := fmt.Errorf("%s-%d", service, filter.HashCode())
	val, ok := f.items.Load(key)
	if ok {
		ce := val.(*cacheEntry)
		if ce.hashCode == entry.HashCode() {
			return ce.entry, nil
		}
	}

	result, err := filter.Do(entry)
	if err != nil {
		return nil, err
	}

	ce := cacheEntry{entry: result, hashCode: entry.HashCode()}
	f.items.Store(key, ce)

	return result, nil
}
