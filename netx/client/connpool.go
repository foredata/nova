package client

import (
	"context"
	"sync"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/netx/discovery"
)

type ConnPool interface {
	Get(ctx context.Context, ins discovery.Instance, tran netx.Tran, opts *CallOptions) (netx.Conn, error)
	Remove(ctx context.Context, ins discovery.Instance)
	Close() error
}

func newConnPool() ConnPool {
	cp := &connpool{
		conns: make(map[string]netx.Conn),
	}
	return cp
}

// connpool 默认使用单独长链接
type connpool struct {
	mux   sync.RWMutex
	conns map[string]netx.Conn
}

// Get 同步阻塞链接
func (c *connpool) Get(ctx context.Context, ins discovery.Instance, tran netx.Tran, opts *CallOptions) (netx.Conn, error) {
	c.mux.RLock()
	conn := c.conns[ins.Addr()]
	if conn != nil {
		c.mux.RUnlock()
		return conn, nil
	}
	c.mux.RUnlock()
	// TODO: use singleflight,目前有问题,相同addr可能会被连接多次?
	conn, err := tran.Dial(ins.Addr())
	if err != nil {
		return nil, err
	}

	c.mux.Lock()
	c.conns[ins.Addr()] = conn
	c.mux.Unlock()
	return conn, nil
}

func (c *connpool) Remove(ctx context.Context, ins discovery.Instance) {
	c.mux.Lock()
	conn := c.conns[ins.Addr()]
	if conn != nil {
		conn.Close()
		delete(c.conns, ins.Addr())
	}
	c.mux.Unlock()
}

func (c *connpool) Close() error {
	c.mux.Lock()
	for _, conn := range c.conns {
		conn.Close()
	}
	c.conns = make(map[string]netx.Conn)
	c.mux.Unlock()
	return nil
}
