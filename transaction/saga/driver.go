package saga

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// some error
var (
	ErrNotFindSaga = errors.New("not find saga")
)

// Action saga执行子事务,需要提供正向执行回调和反向补偿回调
type Action interface {
	Execute(ctx context.Context, params interface{}) error
	Rollback(ctx context.Context, params interface{})
}

// Driver 事务编排分为两种模式Orchestration和Choreography,这里实现了Orchestration模式
// Orchestration属于集中式管理，有一个集中的事务编排者
// Choreography基于事件驱动,没有一个集中的事务编排者，实现更加复杂
// 同一事务使用相同入参，链式调用通过context透传额外上下文参数
type Driver interface {
	// Register 注册事务
	Register(name string, paramType reflect.Type, actions []Action) error
	// Run 执行事务
	Run(ctx context.Context, name string, params interface{}, opts ...RunOption) error
}

func NewDriver(consumer Comsumer, opts ...NewOption) Driver {
	return nil
}

type saga struct {
	Name    string
	Type    reflect.Type
	Actions []Action
}

type driver struct {
	consumer Comsumer
	mgr      Manager
	sagas    map[string]*saga
}

func (d *driver) Register(name string, actions []Action) error {
	return nil
}

// Run 执行事务
func (d *driver) Run(ctx context.Context, name string, params interface{}, opts ...RunOption) error {
	saga := d.sagas[name]
	if saga == nil {
		return ErrNotFindSaga
	}

	o := newRunOptions(opts...)
	// 开启事务
	tx := &Transaction{
		Id:       o.txId,
		Index:    0,
		Rollback: false,
	}

	if err := d.mgr.Start(ctx, tx); err != nil {
		return err
	}

	// 执行事务
	for _, a := range saga.Actions {
		err := a.Execute(ctx, params)
		if err != nil {
			break
		}
	}

	// 更新事务
	d.mgr.Update(ctx, tx)

	return nil
}

// Handle 响应异步补偿事件
func (d *driver) Handle(ctx context.Context, msg *Message) error {
	saga := d.sagas[msg.Name]
	if saga == nil {
		return ErrNotFindSaga
	}

	if msg.TxIndex >= uint(len(saga.Actions)) {
		return fmt.Errorf("saga index overflow, %s, %d", msg.Name, msg.TxIndex)
	}

	var params interface{}
	if saga.Type != nil && len(msg.Params) != 0 {
		p := reflect.New(saga.Type)
		if err := json.Unmarshal(msg.Params, p.Interface()); err != nil {
			return fmt.Errorf("Unmarshal params fail, %+v", err)
		}
		params = p.Interface()
	}

	action := saga.Actions[msg.TxIndex]
	if msg.Rollback {
		action.Rollback(ctx, params)
	} else {
		action.Execute(ctx, params)
	}

	return nil
}
