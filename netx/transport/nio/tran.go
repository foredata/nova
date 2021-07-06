package nio

import (
	"github.com/foredata/nova/netx"
)

// New 创建transport
func New() netx.Tran {
	loop, _ := newLoop()
	t := &nioTran{loop: loop}
	return t
}

// 可参考: https://github.com/mailru/easygo
type nioTran struct {
	netx.BaseTran
	loop *nioLoop
}

func (t *nioTran) String() string {
	return "nio"
}

func (t *nioTran) Listen(addr string, opts ...netx.Option) (netx.Listener, error) {
	o := netx.NewOptions(opts...)
	l, err := o.Listen(addr, o)
	if err != nil {
		return nil, err
	}

	return newListener(l, t, o.Tag)
}

func (t *nioTran) Dial(addr string, opts ...netx.Option) (netx.Conn, error) {
	o := netx.NewOptions(opts...)
	conn := newConn(t, true, o.Tag)
	if o.DialNonBlocking {
		go func() {
			_, _ = t.doDial(conn, addr, o)
		}()
		return conn, nil
	} else {
		return t.doDial(conn, addr, o)
	}
}

func (t *nioTran) doDial(conn *nioConn, addr string, o *netx.Options) (netx.Conn, error) {
	raw, err := o.Dial(addr, o)
	if err == nil {
		err = conn.Open(raw)
	}
	if o.DialCallback != nil {
		o.DialCallback(conn, err)
	}
	return conn, err
}

// defaultListen 默认实现listen,不能使用net.Listen,因为会阻塞调用
// func defaultListen(host string, opts *netx.Options) (net.Listener, error) {
// 	return net.Listen(opts.Network, host)
// }

// // defaultDial 默认实现Dial
// func defaultDial(addr string, opts *netx.Options) (net.Conn, error) {
// 	return net.DialTimeout(opts.Network, addr, opts.DialTimeout)
// }
