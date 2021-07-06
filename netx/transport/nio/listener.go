package nio

import (
	"net"

	"github.com/foredata/nova/netx/transport/nio/netpoll"
)

func newListener(raw net.Listener, tran *nioTran, tag string) (*nioListener, error) {
	l := &nioListener{
		raw:    raw,
		tran:   tran,
		tag:    tag,
		poller: tran.loop.Major(),
	}
	if err := l.Open(); err != nil {
		return nil, err
	}

	return l, nil
}

type nioListener struct {
	poller netpoll.Poller
	raw    net.Listener
	tran   *nioTran
	tag    string
	fd     netpoll.FD
}

func (l *nioListener) Fd() netpoll.FD {
	return l.fd
}

func (l *nioListener) Addr() net.Addr {
	return l.raw.Addr()
}

func (l *nioListener) Open() error {
	fd, err := netpoll.GetNonblockFd(l.raw)
	if err != nil {
		return err
	}
	l.fd = fd

	if err := l.poller.Insert(l, netpoll.EventIn); err != nil {
		return err
	}

	return nil
}

func (l *nioListener) Close() error {
	if l.raw != nil {
		_ = l.poller.Delete(l)
		err := netpoll.Close(l.fd)
		l.raw = nil
		l.fd = 0
		return err
	}
	return nil
}

// OnEvent 处理EventLoop事件回调
func (l *nioListener) OnEvent(events netpoll.Event) {
	for {
		raw, err := l.raw.Accept()
		if err != nil {
			break
		}
		conn := newConn(l.tran, false, l.tag)
		_ = conn.Open(raw)
	}
}
