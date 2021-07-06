package nio

import (
	"net"
	"syscall"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/netx/transport/nio/netpoll"
)

func newConn(tran *nioTran, client bool, tag string) *nioConn {
	conn := &nioConn{poller: tran.loop.Next()}
	conn.Init(tran, client, tag)
	return conn
}

type nioConn struct {
	netx.BaseConn
	conn   net.Conn       // 原始连接
	fd     netpoll.FD     //
	poller netpoll.Poller //
}

func (c *nioConn) Fd() netpoll.FD {
	return c.fd
}

func (c *nioConn) Open(conn net.Conn) error {
	var err error
	c.Lock()
	if c.conn == nil {
		fd, err := netpoll.GetNonblockFd(conn)
		if err == nil {
			c.fd = fd
			err = c.poller.Insert(c, netpoll.EventIn)
		}

		if err == nil {
			c.conn = conn
			c.SetLocalAddr(conn.LocalAddr().String())
			c.SetRemoteAddr(conn.RemoteAddr().String())
			c.SetStatus(netx.OPEN)
		}
	} else {
		err = netx.ErrConnOpened
	}
	c.Unlock()

	if err == nil {
		c.GetChain().HandleOpen(c)
	} else {
		c.GetChain().HandleError(c, err)
	}

	return err
}

func (c *nioConn) Close() error {
	c.Lock()
	if c.IsStatus(netx.OPEN) {
		if c.GetWriter().Empty() {
			// 直接关闭
			c.doClose(nil)
		} else {
			c.SetStatus(netx.CLOSING)
		}
	}
	c.Unlock()
	return nil
}

func (c *nioConn) Send(msg interface{}) error {
	return c.GetChain().HandleWrite(c, msg)
}

func (c *nioConn) Write(w netx.WriterTo) error {
	c.Lock()
	var err error
	status := c.Status()
	switch status {
	case netx.CONNECTING:
		c.GetWriter().Append(w)
	case netx.OPEN:
		c.GetWriter().Append(w)
		err = c.doWrite()
	}
	c.Unlock()
	return err
}

func (c *nioConn) OnEvent(events netpoll.Event) {
	if events.Is(netpoll.EventErr) {
		c.doClose(nil)
		return
	}

	if events.Is(netpoll.EventIn) {
		c.doRead()
	}

	if events.Is(netpoll.EventOut) {
		c.Lock()
		_ = c.doWrite()
		c.Unlock()
	}

}

// doRead 执行读操作,直到不能读为止
func (c *nioConn) doRead() {
	var res error
	for {
		c.Lock()
		if !c.IsStatus(netx.OPEN) {
			c.Unlock()
			break
		}

		n, err := c.GetReadBuffer().ReadFromOnce(c.conn)
		c.Unlock()

		if n > 0 {
			c.GetChain().HandleRead(c, c.GetReadBuffer())
		}

		if n <= 0 || err != nil {
			if err == syscall.EAGAIN || err == syscall.EWOULDBLOCK {
				err = nil
			}
			break
		}
	}

	if res != nil {
		c.doClose(res)
	}
}

// doWrite 执行发送操作,直到不能发送为止
func (c *nioConn) doWrite() error {
	_, err := c.GetWriter().WriteTo(c.conn)
	if err == syscall.EAGAIN {
		err = nil
	}

	if err != nil || (c.IsStatus(netx.CLOSING) && c.GetWriter().Empty()) {
		c.doClose(err)
		return nil
	}

	if c.GetWriter().Empty() {
		return c.poller.Modify(c, netpoll.EventIn)
	} else {
		return c.poller.Modify(c, netpoll.EventInOut)
	}
}

// doClose 关闭socket
func (c *nioConn) doClose(err error) {
	if c.IsStatus(netx.CLOSED) {
		return
	}

	if err != nil {
		c.GetChain().HandleError(c, err)
	}

	c.GetWriter().Clear()
	c.SetStatus(netx.CLOSED)
	if c.conn != nil {
		c.poller.Delete(c)
		c.fd = 0
		c.conn = nil
	}
}
