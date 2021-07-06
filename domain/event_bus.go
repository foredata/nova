package domain

import (
	"context"
	"fmt"
)

const (
	// EventNumMax 单次可产生的最大event数
	EventNumMax = 100
)

// EventHandler EventBus事件处理接口
type EventHandler interface {
	HandleEvent(ctx context.Context, event Event) error
}

// EventHandlerFunc 函数实现EventHandler接口
type EventHandlerFunc func(ctx context.Context, ev Event) error

func (f EventHandlerFunc) HandleEvent(ctx context.Context, ev Event) error {
	return f(ctx, ev)
}

// EventBus 通常用于异步事件处理
type EventBus interface {
	// Publish 发布事件
	Publish(ctx context.Context, agg Aggregate, events []Event) error
	// Subscribe 订阅事件,可通过group分组.同组只会有一个处理
	Subscribe(aggType AggregateType, eventType EventType, group string, handler EventHandler) error
	// Close 会阻塞等待所有任务完成
	Close() error
}

// NewSyncEventBus 创建默认EventBus,同步调用
func NewSyncEventBus() EventBus {
	return &syncEventBus{
		handlers: make(map[string][]eventHandlerWrapper),
	}
}

type eventHandlerWrapper struct {
	EventHandler
	group string
}

// syncEventBus 基于内存实现同步EventBus
type syncEventBus struct {
	// aggType+evtType -> EventHandler
	handlers map[string][]eventHandlerWrapper
}

// Publish 发布事件
func (eb *syncEventBus) Publish(ctx context.Context, agg Aggregate, events []Event) error {
	if len(events) >= EventNumMax {
		return ErrTooManyEvents
	}

	for _, ev := range events {
		key := toUniqueKey(agg.AggregateType(), string(ev.EventType()))
		handlers := eb.handlers[key]
		// 因为是同步调用,要求必须存在handler避免遗漏消息未处理
		if len(handlers) == 0 {
			return ErrNotFoundEventHandler
		}

		for _, h := range handlers {
			if err := h.HandleEvent(ctx, ev); err != nil {
				return err
			}
		}
	}

	return nil
}

// Subscribe 订阅事件
func (eb *syncEventBus) Subscribe(aggType AggregateType, eventType EventType, group string, handler EventHandler) error {
	key := toUniqueKey(aggType, string(eventType))
	handlers := eb.handlers[key]
	// check duplicate
	for _, h := range handlers {
		if h.group == group {
			return fmt.Errorf("%w, aggregate type=%s, eventType =%s ", ErrDuplicateEventHandler, aggType, eventType)
		}
	}

	eb.handlers[key] = append(eb.handlers[key], eventHandlerWrapper{EventHandler: handler, group: group})
	return nil
}

func (eb *syncEventBus) Close() error {
	return nil
}
