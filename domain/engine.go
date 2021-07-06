package domain

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/foredata/nova/encoding"
	"github.com/foredata/nova/sync/lock"
	"github.com/foredata/nova/times"
)

// some error
var (
	ErrInvalidAggregateId     = errors.New("invalid aggregate id")
	ErrInvalidCommandType     = errors.New("invalid command type")
	ErrInvalidEvent           = errors.New("invalid event")
	ErrNotFoundEventHandler   = errors.New("not found event handler")
	ErrNotFoundCommandHandler = errors.New("not found command handler")
	ErrNotSupport             = errors.New("not support")
	ErrDuplicateEventHandler  = errors.New("duplicate event handler")
	ErrTooManyEvents          = errors.New("too many events")
)

// Engine 实现EventSourcing+CQRS
// Handle 处理流程
// 	1:查找Command回调
//	2:基于AggregateID加锁
//	2:加载Aggregate
//	3:加载并执行未处理事件
//	4:执行Command并生成单个Event，用于修改Aggregate状态
//	5:保存Event
//	6:应用Event,并生成异步Event用于Projector
//	7:保存Aggregate
// 关键点:
//	一个Command只会产生一个同步处理Event,Apply该Event会产生N个异步Event用于构建QueryModel索引
// 使用方法:
//	1:创建Aggregate
//	2:创建Command和Event，可以是同一个类,也可以是两个类
//	3:在Aggregate中实现CommandHandler和EventHandler,注意:一定要在Aggregate中实现,而且需要对外暴露，因为通过反射解析
type Engine interface {
	// Register 注册Aggregate,基于反射自动注册CommandHandler和EventHandler,非线程安全,仅初始化注册
	Register(aggregates ...Aggregate)
	// Handle 执行Command,需要由外部明确指定关联的Aggregate信息
	Handle(ctx context.Context, cmd Command, opts ...HandleOption) (Aggregate, error)
	// Subscribe 注册异步Event事件,默认使用同步处理,相同event在同一group下只处理一次,不同group下独立执行
	Subscribe(aggType AggregateType, eventType EventType, group string, handler EventHandler) error
	// Close 关闭退出,内部组件可阻塞等待所有消息处理完后优雅退出
	Close() error
}

// Config 配置信息
type Config struct {
	Locking lock.Locking
	Repo    Repository
	Store   EventStore
	Bus     EventBus
	Codec   encoding.Codec // event编解码方式,默认使用json
}

// NewEngine 通过配置创建Engine,在使用时需要由外部提供必要的依赖
func NewEngine(cfg *Config) Engine {
	if cfg.Codec == nil {
		cfg.Codec = encoding.NewJsonCodec()
	}

	if cfg.Bus == nil {
		cfg.Bus = NewSyncEventBus()
	}

	if cfg.Locking == nil {
		cfg.Locking = lock.NewNoopLocking()
	}

	if cfg.Store == nil {
		cfg.Store = &noopEventStore{}
	}

	if cfg.Repo == nil {
		cfg.Repo = &noopRepository{}
	}

	return &engine{
		locking:     cfg.Locking,
		repo:        cfg.Repo,
		store:       cfg.Store,
		bus:         cfg.Bus,
		codec:       cfg.Codec,
		aggregates:  make(map[string]*sync.Pool),
		cmdHandlers: make(handlerMap),
		evtHandlers: make(handlerMap),
	}
}

type method struct {
	Index int          // method index
	Type  reflect.Type // Command/Event Type
}

// Call 执行Command,
// 成员函数签名为:
// funx(ctx context.Context, cmd xxxCmd)(xxxEvent, error)
// func(ctx context.Context, cmd xxxCmd) ([]Event, error)
func (m *method) Call(obj reflect.Value, ctx reflect.Value, cmd reflect.Value) ([]Event, error) {
	mth := obj.Method(m.Index)
	out := mth.Call([]reflect.Value{ctx, cmd})
	if !out[1].IsNil() {
		return nil, out[1].Interface().(error)
	}

	out0 := out[0]
	if out0.Type().Kind() == reflect.Slice {
		return out0.Interface().([]Event), nil
	}

	return []Event{out0.Interface().(Event)}, nil
}

// Apply 应用事件,并创建QueryEvent,用于构建QueryModel,可以产生0个或多个
// 成员函数签名为:
// func(ctx cotnext.Context, event xxxEvent) error
func (m *method) Apply(obj reflect.Value, ctx reflect.Value, event reflect.Value) error {
	mt := obj.Method(m.Index)
	in := []reflect.Value{ctx, event}
	ot := mt.Call(in)
	if ot[0].IsNil() {
		return nil
	}

	return ot[0].Interface().(error)
}

type handlerMap map[string]*method

func (m handlerMap) Insert(aggType AggregateType, subType string, mth *method) error {
	key := toUniqueKey(aggType, subType)
	if _, ok := m[key]; ok {
		return fmt.Errorf("duplicate method register, %+v, %+v", aggType, subType)
	}
	m[key] = mth
	return nil
}

func (m handlerMap) Find(aggType AggregateType, subType string) *method {
	key := toUniqueKey(aggType, subType)
	return m[key]
}

type engine struct {
	locking     lock.Locking          // Aggregate锁,同一时刻同一ID只能有1个事件在处理
	repo        Repository            // 需要外部提供
	store       EventStore            // 需要外部提供
	bus         EventBus              // 默认同步调用
	codec       encoding.Codec        // event编解码,默认json编码
	aggregates  map[string]*sync.Pool // aggType->aggPool
	cmdHandlers handlerMap            // aggType+cmdType->methodInfo
	evtHandlers handlerMap            // aggType+evtType->methodInfo
}

// Register 注册Aggregate,通过反射解析CommandHandler/EventHandler
// CommandHandler:
//	func(ctx context.Context, cmd xxxCmd) ([]Event, error)
// EventHandler:
//	func(ctx context.Context, evt xxxCmd) error
func (e *engine) Register(aggregates ...Aggregate) {
	for _, agg := range aggregates {
		if _, ok := e.aggregates[string(agg.AggregateType())]; ok {
			panic(fmt.Errorf("duplicate register aggregate, type=%s", agg.AggregateType()))
		}

		rv := reflect.ValueOf(agg)
		rt := rv.Type()

		e.aggregates[string(agg.AggregateType())] = &sync.Pool{
			New: func() interface{} {
				return reflect.New(rt.Elem()).Interface()
			},
		}

		for i := 0; i < rv.NumMethod(); i++ {
			f := rv.Method(i).Type()
			if f.NumIn() != 2 || !isContext(f.In(0)) {
				continue
			}

			if f.NumOut() != 1 && f.NumOut() != 2 {
				continue
			}

			if !isError(f.Out(f.NumOut() - 1)) {
				continue
			}

			in1 := f.In(1)

			if !isCommand(in1) && !isEvent(in1) {
				continue
			}

			mth := &method{Index: i, Type: in1}

			if isEvent(in1) && f.NumOut() == 1 {
				e.evtHandlers.Insert(agg.AggregateType(), string(toEventType(in1)), mth)
				continue
			}

			out0 := f.Out(0)
			switch {
			case isCommand(in1) && (isEventSlice(out0) || isEvent(out0)):
				// CommandHandler
				e.cmdHandlers.Insert(agg.AggregateType(), string(toCommandType(in1)), mth)
			case isEvent(in1) && f.NumOut() == 1:
				// EventHandler
				e.evtHandlers.Insert(agg.AggregateType(), string(toEventType(in1)), mth)
			default:
				panic(fmt.Errorf("invalid handler signature[%+v], please check command/event handler signature", f.Name()))
			}
		}
	}
}

// Handle 执行命令
func (e *engine) Handle(ctx context.Context, cmd Command, opts ...HandleOption) (Aggregate, error) {
	o := newHandleOptions(opts...)
	if o.aggType == "" {
		if agg, ok := cmd.(Aggregatable); ok {
			o.aggType = agg.AggregateType()
			o.aggId = agg.AggregateID()
		} else {
			return nil, fmt.Errorf("invalid aggregate type, %+v", cmd.CommandType())
		}
	}

	if o.aggId == "" {
		return nil, ErrInvalidAggregateId
	}

	aggType := o.aggType
	aggId := o.aggId

	cmdVal := reflect.ValueOf(cmd)

	// find handler
	cmdKey := fmt.Sprintf("%s:%s", aggType, (string)(cmd.CommandType()))
	cmdInfo := e.cmdHandlers[cmdKey]
	if cmdInfo != nil && cmdInfo.Type != cmdVal.Type() {
		return nil, fmt.Errorf("%w, aggType=%s, cmdType=%s, realType=%s", ErrInvalidCommandType, aggType, cmd.CommandType(), cmdInfo.Type.Name())
	}

	pool := e.aggregates[string(aggType)]

	// 加锁,悲观锁,防止多个请求同时操作同一个aggregate
	uniqueKey := fmt.Sprintf("ddd:%s:%s", string(aggType), aggId)
	l, err := e.locking.Acquire(uniqueKey, nil)
	if err != nil {
		return nil, err
	}
	defer l.Unlock()

	agg := pool.Get().(Aggregate)
	// 尝试加载snapshot
	err = e.repo.Load(ctx, agg)
	if err != nil {
		return nil, err
	}

	ctxVal := reflect.ValueOf(ctx)
	aggVal := reflect.ValueOf(agg)

	// 加载尚未处理的Event
	perEvents, err := e.store.Load(ctx, aggType, aggId, agg.Version())
	if err != nil {
		return nil, err
	}

	for _, pev := range perEvents {
		mth := e.evtHandlers.Find(pev.AggregateType(), string(pev.EventType()))
		if mth == nil {
			return nil, fmt.Errorf("%w, aggType=%s, evtType=%s", ErrNotFoundEventHandler, aggType, pev.EventType())
		}

		ev := reflect.New(mth.Type)
		if err := e.codec.Unmarshal(pev.Data(), ev.Interface()); err != nil {
			return nil, fmt.Errorf("decode event fail, %+v", err)
		}

		if err := e.apply(ctx, mth, agg, aggVal, ctxVal, ev, pev.Version()); err != nil {
			return nil, err
		}
	}

	// 执行command,优先使用注册的Handler,然后执行Aggregate实现的Handler,最后尝试直接转换为Event
	// 一个Command只会转换为一个Event
	var events []Event
	if cmdInfo != nil {
		events, err = cmdInfo.Call(aggVal, ctxVal, cmdVal)
	} else if handler, ok := agg.(CommandHandler); ok {
		// 尝试调用通用handler
		events, err = handler.HandleCommand(ctx, cmd)
	} else if ev, ok := cmd.(Event); ok {
		events = []Event{ev}
	} else {
		return nil, fmt.Errorf("%w, aggType=%s, cmdType=%s", ErrNotFoundCommandHandler, aggType, cmd.CommandType())
	}

	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return nil, ErrInvalidEvent
	}

	// store events
	now := times.Now().UnixNano() / int64(time.Millisecond)
	perEvents = perEvents[:0]
	for i, event := range events {
		eventData, err := e.codec.Marshal(event)
		if err != nil {
			return nil, fmt.Errorf("encode event fail, %w", err)
		}
		pe := NewPersistenceEvent(aggType, aggId, event.EventType(), now, agg.Version()+int64(i)+1, eventData)
		perEvents = append(perEvents, pe)
	}

	if err := e.store.Save(ctx, perEvents); err != nil {
		return nil, fmt.Errorf("save events fail, %w", err)
	}

	// apply events
	for _, event := range events {
		version := agg.Version() + 1

		mth := e.evtHandlers.Find(aggType, string(event.EventType()))
		if mth == nil {
			return nil, fmt.Errorf("%w, aggType=%s, evtType=%s", ErrNotFoundEventHandler, aggType, event.EventType())
		}
		if err := e.apply(ctx, mth, agg, aggVal, ctxVal, reflect.ValueOf(event), version); err != nil {
			return nil, err
		}
	}

	// publish events
	if err := e.bus.Publish(ctx, agg, events); err != nil {
		return nil, fmt.Errorf("publish events fail, %w", err)
	}

	// save snapshot
	if err := e.repo.Save(ctx, agg); err != nil {
		return nil, fmt.Errorf("save snapshot fail, %w", err)
	}

	return agg, nil
}

func (e *engine) apply(ctx context.Context, mth *method, agg Aggregate, aggVal, ctxVal, evtVal reflect.Value, version int64) error {
	err := mth.Apply(aggVal, ctxVal, evtVal)
	if err != nil {
		return fmt.Errorf("handle event fail, %+v", err)
	}

	agg.SetVersion(version)

	return nil
}

func (e *engine) Subscribe(aggType AggregateType, eventType EventType, group string, handler EventHandler) error {
	return e.bus.Subscribe(aggType, eventType, group, handler)
}

func (e *engine) Close() error {
	_ = e.bus.Close()
	return nil
}
