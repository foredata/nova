package job

import (
	"context"
)

// Event 事件
type Event interface {
	Type() string
}

// CancelEvent 取消任务事件
type CancelEvent struct {
	InstanceID string // 唯一ID
}

func (ev *CancelEvent) Type() string {
	return "cancel_event"
}

// Store 存储任务数据
type Store interface {
	// Save 更新统计信息
	Save(ctx context.Context, id string, info *Info, job *Job) error
	// Load 加载统计信息
	Load(ctx context.Context, id string) (*Info, error)
	// Delete 删除统计信息
	Delete(ctx context.Context, id string) error
	// IsRunning 通过任务名查询任务是否在运行中
	IsRunning(ctx context.Context, jobName string) (bool, error)
	// Publish 发布事件
	Publish(ctx context.Context, ev Event) error
	// Subscribe 订阅任务状态变更
	Subscribe(ctx context.Context, fn func(ev Event))
}

func newNoopStore() Store {
	return &noopStore{}
}

// noopStore 不存储数据,注意:
type noopStore struct {
}

func (s *noopStore) Save(ctx context.Context, id string, info *Info, job *Job) error {
	return nil
}

func (s *noopStore) Load(ctx context.Context, id string) (*Info, error) {
	return nil, nil
}

func (s *noopStore) Delete(ctx context.Context, id string) error {
	return nil
}

func (s *noopStore) IsRunning(ctx context.Context, jobName string) (bool, error) {
	return false, nil
}

func (s *noopStore) Publish(ctx context.Context, ev Event) error {
	return nil
}

func (s *noopStore) Subscribe(ctx context.Context, fn func(ev Event)) {

}
