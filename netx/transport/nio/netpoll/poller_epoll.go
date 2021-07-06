// +build linux,!noepoll

package netpoll

import (
	"fmt"
	"syscall"
	"unsafe"
)

func newPoller() (Poller, error) {
	p := &epollPoller{}
	if err := p.Open(); err != nil {
		return nil, err
	}

	return p, nil
}

// TODO: 未测试
type epollPoller struct {
	efd    int            // epoll fd
	wakeup *wakeupChannel //
	events []epollEvent   // events
}

func (p *epollPoller) Open() error {
	fd, err := syscall.EpollCreate1(0)
	if err != nil {
		fd, err = syscall.EpollCreate(1024)
		if err != nil {
			return err
		}
	}

	r0, _, e0 := syscall.Syscall(syscall.SYS_EVENTFD2, 0, 0, 0)
	if e0 != 0 {
		_ = syscall.Close(fd)
		return fmt.Errorf("create eventfd fail")
	}

	wc := &wakeupChannel{fd: FD(r0)}
	if err := epollCtlChannel(fd, syscall.EPOLL_CTL_ADD, syscall.EPOLLIN|syscall.EPOLLOUT, wc); err != nil {
		_ = syscall.Close(fd)
		_ = syscall.Close(FD(r0))
		return err
	}

	syscall.CloseOnExec(fd)

	p.efd = fd
	p.wakeup = wc
	p.events = make([]epollEvent, maxEventNum)

	return nil
}

func (p *epollPoller) Close() error {
	var err error
	if p.wakeup != nil {
		if e := syscall.Close(p.wakeup.fd); e != nil {
			err = e
		}
		p.wakeup = nil
	}

	if e := syscall.Close(p.efd); e != nil {
		err = e
	}

	return err
}

func (p *epollPoller) Wakeup() error {
	if p.wakeup != nil {
		_, err := syscall.Write(p.wakeup.fd, []byte{0, 0, 0, 0, 0, 0, 0, 1})
		return err
	}

	return nil
}

func (p *epollPoller) Wait() error {
	n, err := epollWait(p.efd, p.events, 0)
	if err != nil {
		if errno, ok := err.(syscall.Errno); ok && errno.Temporary() {
			return nil
		}
		return err
	}
	for i := 0; i < n; i++ {
		ev := p.events[i]

		channel := *(*Channel)(unsafe.Pointer(&ev.data))
		if channel == nil {
			continue
		}

		if channel.Fd() == p.wakeup.fd {
			continue
		}

		var events Event

		// https://stackoverflow.com/questions/24119072/how-to-deal-with-epollerr-and-epollhup/29206631
		// https://blog.csdn.net/halfclear/article/details/78061771?utm_source=blogxgwz8
		// from libev
		if (ev.events & syscall.EPOLLERR) != 0 {
			events |= EventErr
		}

		if ev.events&(syscall.EPOLLIN|syscall.EPOLLHUP) != 0 {
			events |= EventIn
		}

		if ev.events&(syscall.EPOLLOUT|syscall.EPOLLHUP) != 0 {
			events |= EventOut
		}
		channel.OnEvent(events)
	}

	return nil
}

func (p *epollPoller) Insert(channel Channel, events Event) error {
	mask := uint32(syscall.EPOLLET)
	if events.Is(EventIn) {
		mask |= syscall.EPOLLIN
	}
	if events.Is(EventOut) {
		mask |= syscall.EPOLLOUT
	}

	return epollCtlChannel(p.efd, syscall.EPOLL_CTL_ADD, mask, channel)
}

func (p *epollPoller) Modify(channel Channel, events Event) error {
	mask := uint32(syscall.EPOLLET)
	if events.Is(EventIn) {
		mask |= syscall.EPOLLIN
	}
	if events.Is(EventOut) {
		mask |= syscall.EPOLLOUT
	}

	return epollCtlChannel(p.efd, syscall.EPOLL_CTL_MOD, mask, channel)
}

func (p *epollPoller) Delete(channel Channel) error {
	return epollCtl(p.efd, syscall.EPOLL_CTL_DEL, channel.Fd(), nil)
}

type wakeupChannel struct {
	fd FD
}

func (c *wakeupChannel) Fd() FD {
	return c.fd
}

func (c *wakeupChannel) OnEvent(event Event) {
}

type epollEvent struct {
	events uint32
	data   [8]byte // unaligned uintptr
}

func epollCtlChannel(epfd int, op int, events uint32, channel Channel) error {
	ev := &epollEvent{events: events}
	*(*Channel)(unsafe.Pointer(&ev.data)) = channel
	return epollCtl(epfd, op, channel.Fd(), ev)
}

// EpollCtl implements epoll_ctl.
func epollCtl(epfd int, op int, fd int, event *epollEvent) (err error) {
	_, _, err = syscall.RawSyscall6(syscall.SYS_EPOLL_CTL, uintptr(epfd), uintptr(op), uintptr(fd), uintptr(unsafe.Pointer(event)), 0, 0)
	if err == syscall.Errno(0) {
		err = nil
	}
	return err
}

// epollWait implements epoll_wait.
func epollWait(epfd int, events []epollEvent, msec int) (n int, err error) {
	var r0 uintptr
	var _p0 = unsafe.Pointer(&events[0])
	if msec == 0 {
		r0, _, err = syscall.RawSyscall6(syscall.SYS_EPOLL_WAIT, uintptr(epfd), uintptr(_p0), uintptr(len(events)), 0, 0, 0)
	} else {
		r0, _, err = syscall.Syscall6(syscall.SYS_EPOLL_WAIT, uintptr(epfd), uintptr(_p0), uintptr(len(events)), uintptr(msec), 0, 0)
	}
	if err == syscall.Errno(0) {
		err = nil
	}
	return int(r0), err
}
