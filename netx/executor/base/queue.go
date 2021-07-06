package base

import (
	"sync"

	"github.com/foredata/nova/netx"
)

var gNodePool = sync.Pool{
	New: func() interface{} {
		return &node{}
	},
}

// Queue 任务队列
type Queue struct {
	head *node
	tail *node
	size int
}

type node struct {
	prev *node
	next *node
	task netx.Runnable
}

// Len 返回长度
func (q *Queue) Len() int {
	return q.size
}

// Empty 是否为空
func (q *Queue) Empty() bool {
	return q.size == 0
}

// Swap 交换
func (q *Queue) Swap(o *Queue) {
	*o, *q = *q, *o
}

// Push 末尾追加
func (q *Queue) Push(task netx.Runnable) {
	n := gNodePool.Get().(*node)
	n.task = task
	if q.tail == nil {
		q.head = n
		q.tail = n
	} else {
		n.prev = q.tail
		q.tail.next = n
		q.tail = n
	}
	q.size++
}

// Pop 释放第一个节点,并返回数据
func (q *Queue) Pop() netx.Runnable {
	if q.head != nil {
		t := q.head.task
		n := q.head
		q.head = q.head.next
		if q.head == nil {
			q.tail = nil
		}
		n.next = nil
		return t
	}

	return nil
}

// Process 执行所有task
func (q *Queue) Process() {
	for !q.Empty() {
		task := q.Pop()
		_ = task.Run()
	}
}
