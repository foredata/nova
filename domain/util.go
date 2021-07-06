package domain

import (
	"context"
	"fmt"
	"reflect"
)

// util for reflect CommandHandler/EventHandler
var (
	ctxType = reflect.TypeOf((*context.Context)(nil)).Elem()
	errType = reflect.TypeOf((*error)(nil)).Elem()
	cmdType = reflect.TypeOf((*Command)(nil)).Elem()
	evtType = reflect.TypeOf((*Event)(nil)).Elem()
)

func isContext(t reflect.Type) bool {
	return t.Implements(ctxType)
}

func isError(t reflect.Type) bool {
	return t.Implements(errType)
}

func isCommand(t reflect.Type) bool {
	return t.Implements(cmdType)
}

func isEvent(t reflect.Type) bool {
	return t.Implements(evtType)
}

// isEventSlice check is []Event
func isEventSlice(t reflect.Type) bool {
	return t.Kind() == reflect.Slice && isEvent(t.Elem())
}

func toEventType(t reflect.Type) EventType {
	v := reflect.New(t.Elem()).Interface().(Event)
	return v.EventType()
}

func toCommandType(t reflect.Type) CommandType {
	v := reflect.New(t.Elem()).Interface().(Command)
	return v.CommandType()
}

func toUniqueKey(aggType AggregateType, subType string) string {
	return fmt.Sprintf("%s:%s", string(aggType), subType)
}

func toError(v reflect.Value) error {
	if !v.IsNil() {
		return v.Interface().(error)
	}

	return nil
}
