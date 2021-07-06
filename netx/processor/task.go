package processor

import (
	"sync"
	"time"

	"github.com/foredata/nova/netx"
)

var gSimpleTaskPool = sync.Pool{
	New: func() interface{} {
		return &simpleTask{}
	},
}

func newSimpleTask(conn netx.Conn, packet netx.Packet, callback netx.Callback) *simpleTask {
	t := gSimpleTaskPool.Get().(*simpleTask)
	t.conn = conn
	t.packet = packet
	t.callback = callback
	return t
}

type simpleTask struct {
	conn     netx.Conn
	packet   netx.Packet
	callback netx.Callback
}

func (t *simpleTask) Run() error {
	err := t.callback(t.conn, t.packet)
	gSimpleTaskPool.Put(t)
	return err
}

var gStreamTaskPool = sync.Pool{
	New: func() interface{} {
		return &streamTask{}
	},
}

func newStreamTask(taskId uint64, owner *processor, conn netx.Conn, packet netx.Packet, callback netx.Callback) *streamTask {
	t := gStreamTaskPool.Get().(*streamTask)
	t.taskId = taskId
	t.owner = owner
	t.conn = conn
	t.packet = packet
	t.callback = callback
	t.complete = false
	t.running = false
	t.lastTime = time.Now()
	return t
}

type streamTask struct {
	taskId   uint64
	owner    *processor
	conn     netx.Conn
	packet   netx.Packet
	callback netx.Callback // 消息回调
	complete bool          // 是否处理完成
	running  bool          // 任务是否执行中
	lastTime time.Time     // 上次触发时间
	mux      sync.Mutex    //
}

func (t *streamTask) TryPost(complete bool) error {
	t.mux.Lock()
	running := t.running
	t.running = true
	t.complete = complete
	t.lastTime = time.Now()
	t.mux.Unlock()
	if !running {
		return t.owner.executor.Post(t)
	}

	return nil
}

func (t *streamTask) Run() error {
	// 有并发,需要保证一个消息同时只能在一个线程中执行,并且不能丢失消息
	// 限制要求:handler允许多次执行
	for {
		err := t.callback(t.conn, t.packet)
		t.mux.Lock()
		if err != nil {
			t.running = false
			if !t.complete {
				t.owner.deleteTask(t.taskId)
			}
			gStreamTaskPool.Put(t)
			t.mux.Unlock()
			return err
		} else if t.packet.Body().End() {
			t.running = false
			if t.complete {
				gStreamTaskPool.Put(t)
			}
			t.mux.Unlock()
			break
		} else {
			t.mux.Unlock()
		}
	}
	return nil
}
