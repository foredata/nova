package gpc

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/foredata/nova/netx"
)

func newConn(tran netx.Tran, client bool, tag string) *gpcConn {
	conn := &gpcConn{}
	conn.Init(tran, client, tag)
	conn.cond = sync.NewCond(&conn.Mutex)
	return conn
}

type gpcConn struct {
	netx.BaseConn
	conn net.Conn   // 原始连接
	cond *sync.Cond //
}

func (c *gpcConn) Open(conn net.Conn) error {
	err := c.doOpen(conn)

	if err == nil {
		c.GetChain().HandleOpen(c)
	} else {
		c.GetChain().HandleError(c, err)
	}

	return err
}

func (c *gpcConn) Close() error {
	c.Lock()
	defer c.Unlock()

	if c.IsStatus(netx.OPEN) {
		// 等待数据发送完成
		c.SetStatus(netx.CLOSING)
		c.cond.Signal()
	}
	return nil
}

func (c *gpcConn) Send(msg interface{}) error {
	return c.GetChain().HandleWrite(c, msg)
}

func (c *gpcConn) Write(p netx.WriterTo) error {
	err := c.BaseConn.Write(p)
	if err == nil {
		c.cond.Signal()
	}
	return nil
}

func (c *gpcConn) onError(err error) {
	if err == nil {
		return
	}

	if c.IsStatus(netx.CONNECTING) {
		c.Lock()
		c.GetWriter().Clear()
		c.Unlock()
	}

	c.GetChain().HandleError(c, err)
}

func (c *gpcConn) doOpen(conn net.Conn) error {
	c.Lock()
	defer c.Unlock()
	if c.conn != nil {
		return fmt.Errorf("connection has open, %+w", netx.ErrConnOpened)
	}

	c.conn = conn
	c.SetLocalAddr(conn.LocalAddr().String())
	c.SetRemoteAddr(conn.RemoteAddr().String())
	c.GetReadBuffer().Clear()
	c.SetStatus(netx.OPEN)
	go c.readLoop()
	go c.writeLoop()

	return nil
}

func (c *gpcConn) doClose(err error) {
	if c.IsStatus(netx.CLOSED) {
		return
	}

	c.Lock()
	c.GetWriter().Clear()
	c.SetStatus(netx.CLOSED)
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
	c.Unlock()

	if err != nil {
		c.onError(err)
	}
}

// https://tonybai.com/2015/11/17/tcp-programming-in-golang/
// http://www.zfcode.com/?p=315
func (c *gpcConn) readLoop() {
	conn := c.conn
	rb := c.GetReadBuffer()
	for {
		_, _ = rb.Seek(0, io.SeekEnd)
		_, err := rb.ReadFromOnce(conn)
		if err != nil {
			c.doClose(err)
			c.cond.Signal()
			break
		}

		_, _ = rb.Seek(0, io.SeekStart)
		c.GetChain().HandleRead(c, rb)
	}
}

func (c *gpcConn) writeLoop() {
	wb := c.GetWriter()
	writer := netx.NewWriter()
	for {
		c.Lock()
		for c.isWriteWaiting() {
			c.cond.Wait()
		}
		writer.Swap(wb)
		conn := c.conn
		closed := c.IsStatus(netx.CLOSED)
		c.Unlock()

		if closed {
			break
		}

		if conn != nil && !writer.Empty() {
			_, err := writer.WriteTo(conn)
			if err != nil {
				c.doClose(err)
				break
			}
		}

		if c.IsStatus(netx.CLOSING) {
			c.doClose(nil)
			break
		}
	}
}

func (c *gpcConn) isWriteWaiting() bool {
	return c.IsStatus(netx.CONNECTING) || (c.IsStatus(netx.OPEN) && c.GetWriter().Empty())
}
