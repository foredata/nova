package netx

import (
	"context"
	"sync/atomic"
	"time"
)

var seqIdMax uint32

// NewSeqID 生成新的SeqID,只保证单机唯一且不为零
func NewSeqID() uint32 {
	id := atomic.AddUint32(&seqIdMax, 1)
	if id == 0 {
		return atomic.AddUint32(&seqIdMax, 1)
	}

	return id
}

// CallOptions Call时可选参数
type CallOptions struct {
	DialTimeout time.Duration // 连接超时
	CallTimeout time.Duration //
	Callback    interface{}   // 异步回调函数
	RetryPolicy RetryPolicy   // 重试策略
}

// CallOption .
type CallOption func(*CallOptions)

// RetryPolicy 重试策略,比如基于次数
type RetryPolicy interface {
	Allow(ctx context.Context, req Request, retryCount int) bool
}

// Client 客户端接口
type Client interface {
	Call(ctx context.Context, req Request, opts ...CallOption) (Response, error)
	Close() error
}
