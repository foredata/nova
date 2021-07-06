// +build darwin dragonfly freebsd netbsd openbsd

package netpoll

import (
	"io"
	"syscall"
)

func newPoller() (Poller, error) {
	p := &kqueuePoller{}
	if err := p.Open(); err != nil {
		return nil, err
	}

	return p, nil
}

/*
kevent_t:
– ident: 标记事件的描述符, socketfd, filefd, signal
- filter: 事件的类型, 读事件:EVFILT_READ, 写事件:EVFILT_WRITE, 信号:EVFILT_SIGNAL
- flags: 事件的行为, 对kqueue的操作
	- EV_ENABLE: Permit kevent() to return the event if it is triggered.
	- EV_DISABLE: Disable the event so kevent() will not return it.  The filter itself is not disabled.
	- EV_ADD:添加事件
	- EV_DELETE：删除事件
		 Removes the event from the kqueue.  Events which are
                 attached to file descriptors are automatically deleted on
                 the last close of the descriptor.
	- EV_ONESHOT: 一次性事件,指定了该事件, kevent()返回后, 事件会从kqueue中删除
	- EV_CLEAR: 当事件通知给用户后，事件的状态会被重置,可以用在类似于epoll的ET模式，也可以用在描述符有时会出错的情况
	- EV_EOF:  Filters may set this flag to indicate filter-specific EOF condition.
	- EV_ERROR: See RETURN VALUES below.
	- EV_DISPATCH: Disable the event source immediately after delivery of an event.  See EV_DISABLE above.
	- EV_RECEIPT: This flag is useful for making bulk changes to a kqueue
                 without draining any pending events.  When passed as input,
                 it forces EV_ERROR to always be returned.  When a filter is
                 successfully added the data field will be zero.
*/
// TODO: 使用udata代替map
type kqueuePoller struct {
	kfd      int                // kqueue fd
	kevents  []syscall.Kevent_t // kqueue events
	channels channelMap         // 无法使用udata?
}

func (p *kqueuePoller) Open() error {
	p.channels.Init()
	fd, err := syscall.Kqueue()
	if err != nil {
		return err
	}

	changes := []syscall.Kevent_t{{
		Ident:  0,
		Filter: syscall.EVFILT_USER,
		Flags:  syscall.EV_ADD | syscall.EV_CLEAR,
	}}
	_, err = syscall.Kevent(fd, changes, nil, nil)
	if err != nil {
		_ = syscall.Close(fd)
		return err
	}

	syscall.CloseOnExec(fd)
	p.kfd = fd
	p.kevents = make([]syscall.Kevent_t, maxEventNum)

	return nil
}

func (p *kqueuePoller) Close() error {
	if p.kfd != 0 {
		err := syscall.Close(p.kfd)
		p.kfd = 0
		return err
	}

	return nil
}

func (p *kqueuePoller) Wakeup() error {
	changes := []syscall.Kevent_t{{
		Ident:  0,
		Filter: syscall.EVFILT_USER,
		Fflags: syscall.NOTE_TRIGGER,
	}}
	_, err := syscall.Kevent(p.kfd, changes, nil, nil)
	return err
}

func (p *kqueuePoller) Wait() error {
	n, err := syscall.Kevent(p.kfd, nil, p.kevents, nil)
	if err != nil && err != syscall.EINTR {
		if err == syscall.EBADF {
			return io.EOF
		}
		return err
	}

	for i := 0; i < n; i++ {
		kev := &p.kevents[i]
		if kev.Ident == 0 {
			continue
		}

		// channel := *(*Channel)(unsafe.Pointer(kev.Udata))
		// if channel == nil {
		// 	ch, ok := p.channels.Load(kev.Ident)
		// 	if !ok {
		// 		return fmt.Errorf("kqueue: invalid channel,%+v", kev.Ident)
		// 	}
		// 	channel = ch.(Channel)
		// }
		ch := p.channels.Get(FD(kev.Ident))
		if ch == nil {
			continue
		}
		channel := ch.(Channel)

		var events Event
		switch {
		case kev.Flags&syscall.EV_ERROR != 0:
			events |= EventErr
		case kev.Filter == syscall.EVFILT_READ:
			events |= EventIn
		case kev.Filter == syscall.EVFILT_WRITE:
			events |= EventOut
		}
		channel.OnEvent(events)
	}

	return nil
}

func (p *kqueuePoller) Insert(channel Channel, events Event) error {
	err := p.Modify(channel, events)
	if err == nil {
		p.channels.Add(channel)
	}
	return err
}

func (p *kqueuePoller) Modify(channel Channel, events Event) error {
	changes := [2]syscall.Kevent_t{}

	num := 0
	if events.Is(EventIn) {
		changes[num] = syscall.Kevent_t{
			Ident:  uint64(channel.Fd()),
			Filter: syscall.EVFILT_READ,
			Flags:  syscall.EV_ADD | syscall.EV_CLEAR,
			// Udata:  (*byte)(unsafe.Pointer(&channel)),
		}
		num++
	}

	if events.Is(EventOut) {
		changes[num] = syscall.Kevent_t{
			Ident:  uint64(channel.Fd()),
			Filter: syscall.EVFILT_WRITE,
			Flags:  syscall.EV_ADD | syscall.EV_CLEAR,
			// Udata:  (*byte)(unsafe.Pointer(&channel)),
		}
		num++
	}
	if num == 0 {
		return nil
	}

	_, err := syscall.Kevent(p.kfd, changes[:num], nil, nil)
	return err
}

func (p *kqueuePoller) Delete(channel Channel) error {
	p.channels.Del(channel.Fd())

	changes := [2]syscall.Kevent_t{
		{
			Ident:  uint64(channel.Fd()),
			Filter: syscall.EVFILT_READ,
			Flags:  syscall.EV_DELETE,
			// Udata: (*byte)(unsafe.Pointer(&channel)),
		},
		{
			Ident:  uint64(channel.Fd()),
			Filter: syscall.EVFILT_WRITE,
			Flags:  syscall.EV_DELETE,
			// Udata: (*byte)(unsafe.Pointer(&channel)),
		},
	}
	_, err := syscall.Kevent(p.kfd, changes[:], nil, nil)
	return err
}
