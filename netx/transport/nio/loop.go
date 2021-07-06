package nio

import (
	"io"
	"log"
	"runtime"
	"sync/atomic"
	"syscall"

	"github.com/foredata/nova/netx/transport/nio/netpoll"
)

func newLoop() (*nioLoop, error) {
	l := &nioLoop{}
	l.Init(0)
	return l, nil
}

type nioLoop struct {
	major   netpoll.Poller   // 用于listener
	workers []netpoll.Poller // 用于connection
	index   int32
}

func (l *nioLoop) Init(num int) error {
	if num == 0 {
		num = runtime.NumCPU()
	}

	if num < 1 {
		num = 1
	}

	p, err := newPoller()
	if err != nil {
		return err
	}
	l.major = p

	for i := 0; i < num; i++ {
		p, err := newPoller()
		if err != nil {
			return err
		}
		l.workers = append(l.workers, p)
	}

	return nil
}

func (l *nioLoop) Major() netpoll.Poller {
	return l.major
}

func (l *nioLoop) Next() netpoll.Poller {
	index := atomic.AddInt32(&l.index, 1) % int32(len(l.workers))
	return l.workers[index]
}

func newPoller() (netpoll.Poller, error) {
	p, err := netpoll.New()
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			err := p.Wait()
			if err == io.EOF {
				return
			}
			if err != nil && err != syscall.EINTR {
				log.Printf("poll wait fail,%+v", err)
				break
			}
		}
	}()

	return p, nil
}
