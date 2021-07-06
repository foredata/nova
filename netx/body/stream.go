package body

import (
	"io"
	"sync"

	"github.com/foredata/nova/pkg/bytex"
)

// NewStreamBody 用于接收流式数据,每个buffer会作为一个chunk传输
func NewStreamBody(first bytex.Buffer) Body {
	b := &streamBody{}
	b.cond = sync.NewCond(&b.mux)
	if first != nil && first.Len() != 0 {
		b.Write(first)
	}
	return b
}

func newStreamNode(data bytex.Buffer) *streamNode {
	return &streamNode{data: data}
}

type streamNode struct {
	next *streamNode
	data bytex.Buffer
}

// streamBody 由Buffer单链表组成,用于接收流式消息体,当数据不全时,调用Read方法会阻塞
type streamBody struct {
	mux    sync.Mutex  //
	cond   *sync.Cond  //
	head   *streamNode //
	tail   *streamNode //
	ended  bool        // 标记是否传输完成
	closed bool        // 关闭标识
}

func (b *streamBody) End() bool {
	return b.ended
}

func (b *streamBody) Close() error {
	needNotify := false
	b.mux.Lock()
	if !b.closed {
		for b.head != nil {
			node := b.head
			b.head = b.head.next
			node.next = nil
		}
		needNotify = true
		b.head = nil
		b.tail = nil
		b.ended = true
		b.closed = true
	}

	b.mux.Unlock()
	if needNotify {
		b.cond.Signal()
	}

	return nil
}

// Read 读取数据,如果未结束,则会阻塞
func (b *streamBody) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	b.mux.Lock()
	for b.needWait() {
		b.cond.Wait()
	}

	if b.closed {
		b.mux.Unlock()
		return 0, ErrClosed
	}

	total := len(p)
	count := 0
	for b.head != nil && count < total {
		buff := b.head.data
		n, _ := buff.Read(p[count:])
		count += n
		if buff.Len() == buff.Pos() {
			b.popFront()
		}
	}

	if count > 0 {
		b.mux.Unlock()
		return count, nil
	}

	if b.ended {
		b.mux.Unlock()
		return 0, io.EOF
	}

	return 0, nil
}

func (b *streamBody) ReadFast(blocking bool) (bytex.Buffer, error) {
	b.mux.Lock()

	if b.closed {
		b.mux.Unlock()
		return nil, ErrClosed
	}

	if blocking {
		for b.needWait() {
			b.cond.Wait()
		}
	}

	if b.head != nil {
		data := b.popFront()
		b.mux.Unlock()
		return data, nil
	}

	// 传输完成
	if b.ended {
		b.mux.Unlock()
		return nil, io.EOF
	}

	// 不阻塞且没有数据
	return nil, nil
}

func (b *streamBody) Buffer() (bytex.Buffer, error) {
	return nil, ErrNotSupport
}

func (b *streamBody) needWait() bool {
	return b.head == nil && !b.closed && !b.ended
}

func (b *streamBody) popFront() bytex.Buffer {
	data := b.head.data
	node := b.head
	b.head = node.next
	if b.head == nil {
		b.tail = nil
	}
	node.next = nil
	return data
}

// WriteBuffer 末尾追加数据
func (b *streamBody) Write(data bytex.Buffer) {
	if data == nil {
		return
	}
	b.mux.Lock()
	notify := false
	if !b.ended && !b.closed {
		node := newStreamNode(data)
		if b.head == nil {
			b.head = node
			b.tail = node
		} else {
			b.tail.next = node
			b.tail = node
		}

		notify = true
	}

	b.mux.Unlock()
	if notify {
		b.cond.Signal()
	}
}

func (b *streamBody) Flush() {
	b.mux.Lock()
	notify := false
	if !b.ended && !b.closed {
		b.ended = true
		notify = true
	}
	b.mux.Unlock()
	if notify {
		b.cond.Signal()
	}
}
