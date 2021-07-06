package domain

import "context"

// Repository 用于持久化Aggregate,实现snapshot功能
// 外部通常需要实现Repository接口,至少需要实现version的保存,避免重复publish event
type Repository interface {
	Load(ctx context.Context, agg Aggregate) error
	Save(ctx context.Context, agg Aggregate) error
	Delete(ctx context.Context, aggType AggregateType, aggId string) error
}

// noopRepository 空实现,仅用于测试
type noopRepository struct {
}

func (r *noopRepository) Load(ctx context.Context, agg Aggregate) error {
	return nil
}

func (r *noopRepository) Save(ctx context.Context, agg Aggregate) error {
	return nil
}

func (r *noopRepository) Delete(ctx context.Context, aggType AggregateType, aggId string) error {
	return nil
}
