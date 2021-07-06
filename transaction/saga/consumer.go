package saga

import "context"

// Message .
type Message struct {
	Name     string                 //
	TxId     uint64                 // 事务ID
	TxIndex  uint                   // 事务当前索引
	Rollback bool                   // 是否回滚
	Params   []byte                 // 初始参数
	Context  map[string]interface{} // 调用链透传参数
}

// HandleFunc 回调函数
type HandleFunc func(ctx context.Context, msg *Message) error

// Comsumer 当事务处理失败时,异步补偿事务
type Comsumer interface {
	Register(name string, handler HandleFunc)
}
