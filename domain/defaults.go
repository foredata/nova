package domain

import "context"

var gDefault Engine

// SetDefault 设置全局Engine
func SetDefault(e Engine) {
	gDefault = e
}

// Register 注册Aggregate
func Register(aggregates ...Aggregate) {
	gDefault.Register(aggregates...)
}

// Handle .
func Handle(ctx context.Context, cmd Command, opts ...HandleOption) (Aggregate, error) {
	return gDefault.Handle(ctx, cmd, opts...)
}
