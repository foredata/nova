package gpc

import (
	"github.com/foredata/nova/netx"
)

// New 创建基于goroutine-per-conn模式的netx
func New() netx.Tran {
	return &gpcTran{}
}

// 实现goroutine-per-conn
type gpcTran struct {
	netx.BaseTran
}

func (t *gpcTran) String() string {
	return "net"
}

func (t *gpcTran) Listen(addr string, opts ...netx.Option) (netx.Listener, error) {
	o := netx.NewOptions(opts...)
	l, err := o.Listen(addr, o)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			sock, err := l.Accept()
			if err != nil {
				continue
			}
			conn := newConn(t, false, o.Tag)
			_ = conn.Open(sock)
		}
	}()

	return l, nil
}

func (t *gpcTran) Dial(addr string, opts ...netx.Option) (netx.Conn, error) {
	o := netx.NewOptions(opts...)
	conn := newConn(t, true, o.Tag)

	if o.DialNonBlocking {
		go func() {
			_, _ = t.doDial(addr, o, conn)
		}()

		return conn, nil
	} else {
		return t.doDial(addr, o, conn)
	}
}

func (t *gpcTran) doDial(addr string, o *netx.Options, conn *gpcConn) (netx.Conn, error) {
	sock, err := o.Dial(addr, o)

	if err == nil {
		err = conn.Open(sock)
	}

	if o.DialCallback != nil {
		o.DialCallback(conn, err)
	}

	return conn, err
}
