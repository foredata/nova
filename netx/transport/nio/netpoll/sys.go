package netpoll

import (
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
	"time"
)

// osFile describes an object that has ability to return os.File.
type osFile interface {
	// File returns a copy of object's file descriptor.
	File() (*os.File, error)
}

// osFd 用于判断是否实现Fd接口
type osFd interface {
	Fd() FD
}

// GetFd  get fd from socket
func GetFd(socket interface{}) (FD, error) {
	if i, ok := socket.(osFd); ok {
		return i.Fd(), nil
	}

	if i, ok := socket.(osFile); ok {
		if f, err := i.File(); err == nil {
			return FD(f.Fd()), nil
		} else {
			return 0, err
		}
	}

	return 0, fmt.Errorf("netpoll: bad file descriptor")
}

// GetNonblockFd get fd and set nonblock
func GetNonblockFd(socket interface{}) (FD, error) {
	if i, ok := socket.(osFd); ok {
		return i.Fd(), nil
	}

	if i, ok := socket.(osFile); ok {
		if f, err := i.File(); err == nil {
			if err := syscall.SetNonblock(int(f.Fd()), true); err != nil {
				return 0, err
			}
			return FD(f.Fd()), nil
		} else {
			return 0, err
		}
	}

	return 0, fmt.Errorf("netpoll: bad file descriptor")
}

// SetNonblock .
func SetNonblock(fd FD) error {
	return syscall.SetNonblock(fd, true)
}

// Close .
func Close(fd FD) error {
	return syscall.Close(fd)
}

// Listen 实现net.Listen
func Listen(network, address string) (net.Listener, error) {
	sa, st, err := getSockaddr(network, address)
	if err != nil {
		return nil, err
	}

	// create socket
	fd, err := syscall.Socket(st, syscall.SOCK_STREAM, 0)
	if err != nil {
		return nil, err
	}

	// 不设置close后会产生一段时间的TIME_WAIT
	SetReuseAddr(fd)

	// bind
	if err := syscall.Bind(fd, sa); err != nil {
		return nil, err
	}

	if err := syscall.Listen(fd, syscall.SOMAXCONN); err != nil {
		return nil, err
	}

	return newListener(fd, sa), nil
}

// Dial 实现net.Dial,不支持net.DialTimeout
func Dial(network, address string) (net.Conn, error) {
	sa, st, err := getSockaddr(network, address)
	if err != nil {
		return nil, err
	}

	fd, err := syscall.Socket(st, syscall.SOCK_STREAM, 0)
	if err != nil {
		return nil, err
	}

	if err := syscall.Connect(fd, sa); err != nil {
		return nil, err
	}

	return newConn(fd, sa), nil
}

// DialTimeout 实现net.DialTimeout，仅签名保持一致,不支持Timeout
func DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	return Dial(network, address)
}

func getSockaddr(network, address string) (syscall.Sockaddr, int, error) {
	addr, err := net.ResolveTCPAddr(network, address)
	if err != nil {
		return nil, -1, err
	}

	switch network {
	case "tcp", "tcp4":
		var sa4 syscall.SockaddrInet4
		sa4.Port = addr.Port
		copy(sa4.Addr[:], addr.IP.To4())
		return &sa4, syscall.AF_INET, nil
	case "tcp6":
		var sa6 syscall.SockaddrInet6
		sa6.Port = addr.Port
		copy(sa6.Addr[:], addr.IP.To16())
		if addr.Zone != "" {
			ifi, err := net.InterfaceByName(addr.Zone)
			if err != nil {
				return nil, -1, err
			}
			sa6.ZoneId = uint32(ifi.Index)
		}
		return &sa6, syscall.AF_INET6, nil
	default:
		return nil, -1, errors.New("Unknown network type " + network)
	}
}

// getNetAddr returns a go/net friendly address
func getNetAddr(sa syscall.Sockaddr) net.Addr {
	var a net.Addr
	switch sa := sa.(type) {
	case *syscall.SockaddrInet4:
		a = &net.TCPAddr{
			IP:   append([]byte{}, sa.Addr[:]...),
			Port: sa.Port,
		}
	case *syscall.SockaddrInet6:
		var zone string
		if sa.ZoneId != 0 {
			if ifi, err := net.InterfaceByIndex(int(sa.ZoneId)); err == nil {
				zone = ifi.Name
			}
		}
		if zone == "" && sa.ZoneId != 0 {
		}
		a = &net.TCPAddr{
			IP:   append([]byte{}, sa.Addr[:]...),
			Port: sa.Port,
			Zone: zone,
		}
	case *syscall.SockaddrUnix:
		a = &net.UnixAddr{Net: "unix", Name: sa.Name}
	}
	return a
}
