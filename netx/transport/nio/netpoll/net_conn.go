package netpoll

import (
	"net"
	"syscall"
	"time"
)

func newConn(fd FD, remote syscall.Sockaddr) net.Conn {
	sa, err := syscall.Getsockname(fd)
	var local net.Addr
	if err == nil {
		local = getNetAddr(sa)
	} else {
		local = &net.TCPAddr{}
	}
	return &netConn{fd: fd, local: local, remote: getNetAddr(remote)}
}

// netConn 实现net.Conn接口,同时实现Fd()
type netConn struct {
	fd     FD
	local  net.Addr
	remote net.Addr
}

func (c *netConn) Fd() FD {
	return c.fd
}

func (c *netConn) Read(p []byte) (int, error) {
	// TODO: Handle EINTR?
	return syscall.Read(c.fd, p)
}

func (c *netConn) Write(p []byte) (int, error) {
	return syscall.Write(c.fd, p)
}

func (c *netConn) Close() error {
	return syscall.Close(c.fd)
}

func (c *netConn) LocalAddr() net.Addr {
	return c.local
}

func (c *netConn) RemoteAddr() net.Addr {
	return c.remote
}

func (c *netConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *netConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *netConn) SetWriteDeadline(t time.Time) error {
	return nil
}
