package netpoll

import (
	"net"
	"syscall"
)

func newListener(fd FD, addr syscall.Sockaddr) net.Listener {
	return &netListener{fd: fd, addr: getNetAddr(addr)}
}

// netListener 实现net.Listener接口,同时额外提供Fd
type netListener struct {
	fd   FD
	addr net.Addr
}

func (l *netListener) Fd() FD {
	return l.fd
}

func (l *netListener) Addr() net.Addr {
	return l.addr
}

func (l *netListener) Accept() (net.Conn, error) {
	nfd, sa, err := syscall.Accept(l.fd)
	if err != nil {
		return nil, err
	}

	if err := SetNonblock(nfd); err != nil {
		err = syscall.Close(nfd)
		return nil, err
	}

	return newConn(nfd, sa), nil
}

func (l *netListener) Close() error {
	return syscall.Close(l.fd)
}
