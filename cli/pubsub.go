package cli

import "context"

// SubscribeFunc 订阅回调函数
type SubscribeFunc func(ctx context.Context, args []string) (interface{}, error)

// PubSub 发布订阅接口，这里没有约定topic，由实现方确定，比如使用servie_name:admin等
type PubSub interface {
	Publish(ctx context.Context, args []string)
	Subscribe(fn SubscribeFunc)
}

// localPubsub 默认的pubsub，只能本进程执行，不能跨进程pubsub
type localPubsub struct {
	callback SubscribeFunc
}

func (ps *localPubsub) Publish(ctx context.Context, args []string) {
	ps.callback(ctx, args)
}

func (ps *localPubsub) Subscribe(fn SubscribeFunc) {
	ps.callback = fn
}
