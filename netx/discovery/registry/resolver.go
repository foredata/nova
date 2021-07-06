package registry

import (
	"context"
	"sync"

	"github.com/foredata/nova/netx/discovery"
	"github.com/foredata/nova/netx/registry"
)

// New .
func New(reg registry.Registry) discovery.Resolver {
	watcher, _ := reg.Watch(context.Background())
	r := &resolver{
		reg:     reg,
		watcher: watcher,
		nodes:   make(map[string]*registry.Service),
		caches:  make(map[string]*discovery.Result),
	}

	if watcher != nil {
		go r.Watch()
	}

	return r
}

type resolver struct {
	mux     sync.RWMutex
	reg     registry.Registry
	watcher registry.Watcher
	nodes   map[string]*registry.Service // 所有节点,nodeId->service,只有关注的服务才会注册
	caches  map[string]*discovery.Result // 所有服务,name->entry
}

func (r *resolver) Name() string {
	return "registry"
}

func (r *resolver) Resolve(ctx context.Context, service string) (*discovery.Result, error) {
	r.mux.RLock()
	result := r.caches[service]
	if result != nil {
		r.mux.RUnlock()
		return result, nil
	}
	r.mux.RUnlock()

	// 不存在,注册监听
	r.mux.Lock()

	// double check
	result = r.caches[service]
	if result != nil {
		r.mux.Unlock()
		return result, nil
	}

	services, err := r.reg.Get(ctx, service)
	if err != nil {
		r.mux.Unlock()
		return nil, err
	}

	var ins []discovery.Instance
	for _, svc := range services {
		node := svc.Nodes[0]
		r.nodes[node.ID] = svc
		ins = append(ins, toInstance(svc))
	}

	result = discovery.NewResult(ins)
	r.caches[service] = result
	r.mux.Unlock()

	return result, nil
}

func (r *resolver) Close() error {
	return nil
}

func (r *resolver) Watch() {
	for {
		event, err := r.watcher.Next()
		if err != nil {
			continue
		}
		r.mux.Lock()
		node := r.nodes[event.Id]
		switch event.Type {
		case registry.EventDelete:
			if node != nil {
				res := r.caches[node.Name]
				res.Remove(event.Id)
				delete(r.nodes, event.Id)
				// trigger callback?
			}
		case registry.EventCreate, registry.EventUpdate:
			name := event.Service.Name
			res := r.caches[name]
			if res != nil {
				// 仅关注注册过的,未注册的会被忽略掉
				res.Upsert(toInstance(event.Service))
				r.nodes[event.Id] = event.Service
			}
		}
		r.mux.Unlock()
	}
}

// 定期校正?
func (r *resolver) Adjust() {

}

func toInstance(svc *registry.Service) discovery.Instance {
	node := svc.Nodes[0]
	return discovery.NewInstance(node.ID, node.Addr, 0, node.Metadata)
}
