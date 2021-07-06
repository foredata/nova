package netx

import (
	"io"
	"sync"
)

var gNetNodePool = sync.Pool{
	New: func() interface{} {
		return &wnode{}
	},
}

type wnode struct {
	prev *wnode
	next *wnode
	data WriterTo
}

func NewWriter() NetWriter {
	return NetWriter{}
}

// NetWriter 用于队列存储WriterTo,便于单独协程中发送消息
type NetWriter struct {
	head *wnode
	tail *wnode
}

// Empty 判断是否为空
func (w *NetWriter) Empty() bool {
	return w.head == nil
}

// Swap 交换Writer
func (w *NetWriter) Swap(o *NetWriter) {
	*w, *o = *o, *w
}

// Clear 释放所有节点
func (w *NetWriter) Clear() {
	n := w.head
	for n != nil {
		t := n
		n = n.next
		_ = t.data.Close()
		gNetNodePool.Put(t)
	}
	w.head = nil
	w.tail = nil
}

// Append 加入队列
func (w *NetWriter) Append(data WriterTo) {
	node := gNetNodePool.Get().(*wnode)
	node.data = data
	if w.tail == nil {
		w.head = node
		w.tail = node
	} else {
		node.prev = w.tail
		w.tail.next = node
		w.tail = node
	}
}

// WriteTo 将所有数据写入目标
func (w *NetWriter) WriteTo(dst io.Writer) (int64, error) {
	var count int64
	node := w.head
	for node != nil {
		n, err := node.data.WriteTo(dst)
		count += n
		if err != nil {
			node.prev = nil
			w.head = node
			return count, err
		}
		_ = node.data.Close()
		t := node
		node = node.next
		t.next = nil
		t.prev = nil
		gNetNodePool.Put(t)
	}

	w.head = nil
	w.tail = nil
	return count, nil
}
