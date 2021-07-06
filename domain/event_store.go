package domain

import (
	"context"
)

// EventStore 用于存储Event,外部必须提供
type EventStore interface {
	// Save appends all events in the event stream to the store.
	// 如果有多个events,底层存储需要事务保证全部成功
	Save(ctx context.Context, events []PersistenceEvent) error

	// Load loads all events for the aggregate id from the store.
	Load(ctx context.Context, aggType AggregateType, aggId string, fromVersion int64) ([]PersistenceEvent, error)
}

type noopEventStore struct {
}

func (es *noopEventStore) Save(ctx context.Context, events []PersistenceEvent) error {
	return nil
}

func (es *noopEventStore) Load(ctx context.Context, aggType AggregateType, aggId string, fromVersion int64) ([]PersistenceEvent, error) {
	return nil, nil
}

// PersistenceEvent 需要持久化事件
type PersistenceEvent interface {
	AggregateType() AggregateType //
	AggregateID() string          //
	EventType() EventType         //
	Timestamp() int64             // millisecond
	Version() int64               //
	Data() []byte                 // 序列化后数据
}

// NewPersistenceEvent create default persistence event
func NewPersistenceEvent(aggType AggregateType, aggId string, eventType EventType, timestamp int64, version int64, data []byte) PersistenceEvent {
	return &persistenceEvent{
		aggType:   aggType,
		aggId:     aggId,
		eventType: eventType,
		timestamp: timestamp,
		version:   version,
		data:      data,
	}
}

type persistenceEvent struct {
	aggType   AggregateType
	aggId     string
	eventType EventType
	timestamp int64
	version   int64
	data      []byte
}

func (pe *persistenceEvent) AggregateType() AggregateType {
	return pe.aggType
}

func (pe *persistenceEvent) AggregateID() string {
	return pe.aggId
}

func (pe *persistenceEvent) EventType() EventType {
	return pe.eventType
}

func (pe *persistenceEvent) Timestamp() int64 {
	return pe.timestamp
}

func (pe *persistenceEvent) Version() int64 {
	return pe.version
}

func (pe *persistenceEvent) Data() []byte {
	return pe.data
}
