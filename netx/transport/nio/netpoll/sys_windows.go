// +build windows

package netpoll

import (
	"syscall"
	"time"
	"unsafe"
)

type FD = syscall.Handle

func SetReuseAddr(fd FD) {
	syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
}

func goSelect(n int, r, w, e *fdSet, timeout time.Duration) (int, error) {
	var timeval *syscall.Timeval
	if timeout >= 0 {
		t := syscall.NsecToTimeval(timeout.Nanoseconds())
		timeval = &t
	}

	return _select(n, r, w, e, timeval)
}

const fdset_SIZE = 64

type fdSet struct {
	fd_count uint
	fd_array [fdset_SIZE]uintptr
}

// Set adds the fd to the set
func (fds *fdSet) Set(fd uintptr) {
	var i uint
	for i = 0; i < fds.fd_count; i++ {
		if fds.fd_array[i] == fd {
			break
		}
	}
	if i == fds.fd_count {
		if fds.fd_count < fdset_SIZE {
			fds.fd_array[i] = fd
			fds.fd_count++
		}
	}
}

// Clear remove the fd from the set
func (fds *fdSet) Clear(fd uintptr) {
	var i uint
	for i = 0; i < fds.fd_count; i++ {
		if fds.fd_array[i] == fd {
			for i < fds.fd_count-1 {
				fds.fd_array[i] = fds.fd_array[i+1]
				i++
			}
			fds.fd_count--
			break
		}
	}
}

// IsSet check if the given fd is set
func (fds *fdSet) IsSet(fd uintptr) bool {
	if isset, err := __WSAFDIsSet(syscall.Handle(fd), fds); err == nil && isset != 0 {
		return true
	}
	return false
}

// Zero empties the Set
func (fds *fdSet) Zero() {
	fds.fd_count = 0
}

var _ unsafe.Pointer

var (
	modws2_32 = syscall.NewLazyDLL("ws2_32.dll")

	procselect       = modws2_32.NewProc("select")
	proc__WSAFDIsSet = modws2_32.NewProc("__WSAFDIsSet")
)

func _select(nfds int, readfds *fdSet, writefds *fdSet, exceptfds *fdSet, timeout *syscall.Timeval) (total int, err error) {
	r0, _, e1 := syscall.Syscall6(procselect.Addr(), 5, uintptr(nfds), uintptr(unsafe.Pointer(readfds)), uintptr(unsafe.Pointer(writefds)), uintptr(unsafe.Pointer(exceptfds)), uintptr(unsafe.Pointer(timeout)), 0)
	total = int(r0)
	if total == 0 {
		if e1 != 0 {
			err = error(e1)
		}
	}
	return
}

func __WSAFDIsSet(handle syscall.Handle, fdset *fdSet) (isset int, err error) {
	r0, _, e1 := syscall.Syscall(proc__WSAFDIsSet.Addr(), 2, uintptr(handle), uintptr(unsafe.Pointer(fdset)), 0)
	isset = int(r0)
	if isset == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}
