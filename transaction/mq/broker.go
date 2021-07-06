package mq

import "context"

type Message struct {
}

// https://zhuanlan.zhihu.com/p/249233648

// 基于消息队列的事务
type Broker interface {
	Prepare(ctx context.Context, msg *Message) error
	Commit()
	Rollback()
}

type Status uint8

const (
	StatusUnknow   Status = 0 // 未知状态,需要继续询问
	StatusCommit   Status = 1 // 允许订阅方消费该消息
	StatusRollback Status = 2 // 消息将被丢弃不允许消费
)

// Checker 回查线索
type Checker interface {
	Check(ctx context.Context, msg *Message) Status
}

// 使用方式
func onOrder(ctx context.Context) {
	var b Broker
	b.Prepare(ctx, &Message{})

	var err error
	if err != nil {
		b.Rollback()
	} else {
		b.Commit()
	}
}
