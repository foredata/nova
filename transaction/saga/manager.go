package saga

import "context"

// Transaction 事务信息
type Transaction struct {
	Id       int64                  // 事务ID
	Index    int                    // 当前索引
	Rollback bool                   // 是否补偿回滚
	Params   interface{}            // 初始参数
	Context  map[string]interface{} // 链路透传参数
	Result   int                    // 执行结果
}

// Manager 用于管理事务
type Manager interface {
	// Register 注册事务
	Register()
	// Start 启动事务
	Start(ctx context.Context, tx *Transaction) error
	// Update 更新事务
	Update(ctx context.Context, tx *Transaction) error
}
