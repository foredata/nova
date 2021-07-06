package gopool

import (
	"sync/atomic"
	"unsafe"
)

func newQueue() *lfQueue {
	return &lfQueue{}
}

// https://github.com/smallnest/queue/blob/master/lockfree_queue.go
type lfQueue struct {
	head unsafe.Pointer
	tail unsafe.Pointer
}

type lfNode struct {
	value interface{}
	next  unsafe.Pointer
}

// Enqueue .
func (q *lfQueue) Enqueue(v interface{}) {
	n := &lfNode{value: v}
	for {
		tail := load(&q.tail)
		next := load(&tail.next)
		if tail == load(&q.tail) { // are tail and next consistent?
			if next == nil {
				if cas(&tail.next, next, n) {
					cas(&q.tail, tail, n) // Enqueue is done.  try to swing tail to the inserted node
					return
				}
			} else { // tail was not pointing to the last node
				// try to swing Tail to the next node
				cas(&q.tail, tail, next)
			}
		}
	}
}

// Dequeue removes and returns the value at the head of the queue.
// It returns nil if the queue is empty.
func (q *lfQueue) Dequeue() interface{} {
	for {
		head := load(&q.head)
		tail := load(&q.tail)
		next := load(&head.next)
		if head == load(&q.head) { // are head, tail, and next consistent?
			if head == tail { // is queue empty or tail falling behind?
				if next == nil { // is queue empty?
					return nil
				}
				// tail is falling behind.  try to advance it
				cas(&q.tail, tail, next)
			} else {
				// read value before CAS otherwise another dequeue might free the next node
				v := next.value
				if cas(&q.head, head, next) {
					return v // Dequeue is done.  return
				}
			}
		}
	}
}

func load(p *unsafe.Pointer) (n *lfNode) {
	return (*lfNode)(atomic.LoadPointer(p))
}

func cas(p *unsafe.Pointer, old, new *lfNode) (ok bool) {
	return atomic.CompareAndSwapPointer(p, unsafe.Pointer(old), unsafe.Pointer(new))
}
