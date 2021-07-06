package static

import (
	"context"
	"fmt"

	"github.com/foredata/nova/netx/discovery"
)

func New() discovery.Resolver {
	r := &resolver{}
	return r
}

type resolver struct {
}

func (r *resolver) Name() string {
	return "static"
}

func (r *resolver) Resolve(ctx context.Context, addr string) (*discovery.Result, error) {
	if addr == "" {
		return nil, fmt.Errorf("empty addr")
	}

	instances := []discovery.Instance{
		discovery.NewInstance("", addr, 0, nil),
	}
	return discovery.NewResult(instances), nil
}

func (r *resolver) Close() error {
	return nil
}
