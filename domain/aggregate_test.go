package domain

import (
	"context"
	"fmt"
	"testing"
)

const (
	aggTypeOrder      = "order"
	orderCreatedEvent = "created"
)

// 方式一:command和event分开定义
type createCommand struct {
	OrderId string
	Goods   []string
}

func (c *createCommand) CommandType() CommandType {
	return "createCommand"
}

type createdEvent struct {
	OrderId string
	Goods   []string
}

func (e *createdEvent) EventType() EventType {
	return orderCreatedEvent
}

// 可以自动生成Command代码
// @Aggregate(order,OrderId)
type orderAggregate struct {
	AggregateBase
	OrderId string
	Goods   []string
}

func (a *orderAggregate) AggregateType() AggregateType {
	return aggTypeOrder
}

func (a *orderAggregate) AggregateID() string {
	return a.OrderId
}

func (a *orderAggregate) OnCreateCmd(ctx context.Context, cmd *createCommand) (*createdEvent, error) {
	fmt.Printf("orderAggregate:OnCreateCmd, %+v\n", cmd)
	return &createdEvent{OrderId: cmd.OrderId, Goods: cmd.Goods}, nil
}

func (a *orderAggregate) OnCreateEvent(ctx context.Context, ev *createdEvent) error {
	fmt.Printf("orderAggregate:OnCreateEvt, %+v\n", ev)
	a.OrderId = ev.OrderId
	a.Goods = ev.Goods
	return nil
}

func TestOrderAggregate(t *testing.T) {
	e := NewEngine(&Config{})
	e.Register(&orderAggregate{})

	// 测试Subscribe事件
	e.Subscribe(aggTypeOrder, orderCreatedEvent, "", EventHandlerFunc(func(ctx context.Context, ev Event) error {
		t.Logf("consume event, %+v", ev)
		return nil
	}))

	// 测试Cmd&Event分离
	cmd := &createCommand{
		OrderId: "111",
		Goods:   []string{"book"},
	}

	if agg, err := e.Handle(context.Background(), cmd, WithAggregate(aggTypeOrder, cmd.OrderId)); err != nil {
		t.Error(err)
	} else {
		t.Log(agg)
	}
}
