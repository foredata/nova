package processor

import (
	"errors"
	"sync"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/netx/body"
)

var (
	ErrNotFoundHandler = errors.New("not found handler")
)

// New .
func New(executor netx.Executor, provider Provider) netx.Processor {
	p := &processor{
		executor: executor,
		provider: provider,
		tasks:    make(map[uint64]*streamTask),
	}
	return p
}

// Provider 查询回调函数
type Provider interface {
	Find(pkt netx.Packet) netx.Callback
}

type processor struct {
	provider Provider
	executor netx.Executor
	tasks    map[uint64]*streamTask
	mux      sync.RWMutex
}

func (p *processor) Process(conn netx.Conn, frame netx.Frame) error {
	taskId := (uint64(conn.ID()) << 32) | uint64(frame.StreamID())
	if frame.Type() == netx.FrameTypeHeader {
		packet := newPacket(frame)
		callback := p.provider.Find(packet)
		if callback == nil {
			return ErrNotFoundHandler
		}

		if frame.EndFlag() {
			t := newSimpleTask(conn, packet, callback)
			return p.executor.Post(t)
		} else {
			t := newStreamTask(taskId, p, conn, packet, callback)
			p.addTask(t)
			return p.executor.Post(t)
		}
	} else {
		p.mux.RLock()
		task := p.tasks[taskId]
		p.mux.RUnlock()
		if frame.EndFlag() {
			p.deleteTask(taskId)
		}
		if task != nil {
			return task.TryPost(frame.EndFlag())
		}
	}
	return nil
}

func (p *processor) addTask(t *streamTask) {
	p.mux.Lock()
	p.tasks[t.taskId] = t
	p.mux.Unlock()
}

func (p *processor) deleteTask(taskId uint64) {
	p.mux.Lock()
	delete(p.tasks, taskId)
	p.mux.Unlock()
}

func newPacket(f netx.Frame) netx.Packet {
	p := netx.NewPacket()
	bd := body.New(f.Payload(), !f.EndFlag())
	p.SetIdentifier(f.Identifier())
	p.SetHeader(f.Header())
	p.SetBody(bd)
	return p
}
