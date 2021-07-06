package netx

import "net"

// BaseTran basic netx
type BaseTran struct {
	chain     FilterChain
	listeners []net.Listener
}

// GetChain ...
func (t *BaseTran) GetChain() FilterChain {
	return t.chain
}

// SetChain ...
func (t *BaseTran) SetChain(chain FilterChain) {
	if chain != nil {
		t.chain = chain
	}
}

// AddFilters ...
func (t *BaseTran) AddFilters(filters ...Filter) {
	if t.chain == nil {
		t.chain = NewFilterChain()
	}

	t.chain.AddLast(filters...)
}

// Close ...
func (t *BaseTran) Close() error {
	var err error
	for _, l := range t.listeners {
		if e := l.Close(); e != nil {
			err = e
		}
	}

	t.listeners = nil

	return err
}

// AddListener ...
func (t *BaseTran) AddListener(l net.Listener) {
	t.listeners = append(t.listeners, l)
}
