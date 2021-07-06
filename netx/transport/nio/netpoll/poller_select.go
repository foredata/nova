// +build windows

package netpoll

import (
	"os"
	"syscall"
)

func newPoller() (Poller, error) {
	p := &selectPoller{}
	if err := p.Open(); err != nil {
		return nil, err
	}

	return p, nil
}

// TODO: 未测试
type selectPoller struct {
	fdmax    FD
	fds      []FD
	rset     fdSet
	wset     fdSet
	eset     fdSet
	pr       *os.File
	pw       *os.File
	channels channelMap
}

func (p *selectPoller) Open() error {
	p.channels.Init()
	p.fdmax = -1
	p.rset.Zero()
	p.wset.Zero()
	p.eset.Zero()
	r, w, err := os.Pipe()
	if err != nil {
		return err
	}
	p.pr = r
	p.pw = w
	p.rset.Set(r.Fd())
	p.fdmax = FD(r.Fd())
	p.fds = append(p.fds, FD(r.Fd()))

	return nil
}

func (p *selectPoller) Close() error {
	e1 := p.pr.Close()
	e2 := p.pw.Close()
	if e1 != nil {
		return e1
	}
	return e2
}

func (p *selectPoller) Wakeup() error {
	_, err := p.pw.Write([]byte("0"))
	return err
}

func (p *selectPoller) Wait() error {
	p.eset.Zero()
	_, err := goSelect(int(p.fdmax+1), &p.rset, &p.wset, &p.eset, -1)
	if err != nil {
		if errno, ok := err.(syscall.Errno); ok && errno.Temporary() {
			return nil
		}

		return err
	}

	// test wakeup
	if p.rset.IsSet(p.pr.Fd()) {
		// drain all
		bytes := [64]byte{}
		for {
			_, err := p.pr.Read(bytes[:64])
			if err != nil {
				if errno, ok := err.(syscall.Errno); ok && errno.Temporary() {
					continue
				}
				break
			}
		}
	}

	// for loop
	for _, fd := range p.fds {
		var events Event
		if p.eset.IsSet(uintptr(fd)) {
			events |= EventErr
			continue
		}

		if p.rset.IsSet(uintptr(fd)) {
			events |= EventIn
		}
		if p.wset.IsSet(uintptr(fd)) {
			events |= EventOut
		}

		if events != 0 {
			ch := p.channels.Get(fd)
			if ch != nil {
				ch.OnEvent(events)
			}
		}
	}

	return nil
}

func (p *selectPoller) Insert(channel Channel, events Event) error {
	fd := channel.Fd()
	if fd >= p.fdmax {
		p.fdmax = fd
	}
	p.rset.Set(uintptr(fd))
	p.fds = append(p.fds, fd)

	p.channels.Add(channel)
	return nil
}

func (p *selectPoller) Modify(channel Channel, events Event) error {
	return nil
}

func (p *selectPoller) Delete(channel Channel) error {
	fd := channel.Fd()
	p.rset.Clear(uintptr(fd))
	p.wset.Clear(uintptr(fd))

	fdmax := FD(0)
	index := -1
	for i, f := range p.fds {
		if f == fd {
			index = i
		} else if f > fdmax {
			fdmax = f
		}
	}

	p.fdmax = fdmax
	if index != -1 {
		p.fds = append(p.fds[:index], p.fds[index+1:]...)
	}
	p.channels.Del(channel.Fd())

	return nil
}
