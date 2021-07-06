package dns

import (
	"context"

	"github.com/foredata/nova/netx/discovery"
)

// New .
func New() discovery.Resolver {
	r := &resolver{}
	return r
}

type resolver struct {
}

func (r *resolver) Name() string {
	return "dns"
}

func (r *resolver) Resolve(ctx context.Context, service string) (*discovery.Result, error) {
	return nil, nil
}

func (r *resolver) Close() error {
	return nil
}
