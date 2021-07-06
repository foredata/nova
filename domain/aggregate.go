package domain

import (
	"context"
)

// State Aggregate状态,通常用于校验数据是否合法
type State uint

const (
	StateInvalid State = 0 // 初始状态
	StateCreated           // 首次创建
	StateUpdated           // 更新过
	StateDeleted           // 被删除
)

// AggregateType Aggregate类型
type AggregateType string

// Aggregatable 可被聚合对象,常用于Command提供必要Aggregate信息
type Aggregatable interface {
	AggregateType() AggregateType
	AggregateID() string
}

// Aggregate 聚合对象
type Aggregate interface {
	AggregateType() AggregateType
	AggregateID() string
	Version() int64
	SetVersion(v int64)
	State() State
	SetState(s State)
	IsState(s State) bool
	IsAlive() bool // 是否可用,state !=StateInvalid && state != StateDeleted
}

type CommandType string

// Command .
type Command interface {
	CommandType() CommandType
}

// CommandHandler Command处理接口,仅用于通用处理command
type CommandHandler interface {
	HandleCommand(ctx context.Context, cmd Command) ([]Event, error)
}

type EventType string

// Event event
type Event interface {
	EventType() EventType
}

type AggregateBase struct {
	version int64
	state   State
}

func (a *AggregateBase) Version() int64 {
	return a.version
}

func (a *AggregateBase) SetVersion(v int64) {
	a.version = v
}

func (a *AggregateBase) State() State {
	return a.state
}

func (a *AggregateBase) SetState(s State) {
	a.state = s
}

func (a *AggregateBase) IsState(s State) bool {
	return a.state == s
}

func (a *AggregateBase) IsAlive() bool {
	return a.state != StateInvalid && a.state != StateDeleted
}
