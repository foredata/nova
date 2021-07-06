package reflects

import (
	"context"
	"reflect"
)

var (
	ctxType = reflect.TypeOf((*context.Context)(nil)).Elem()
	errType = reflect.TypeOf((*error)(nil)).Elem()
)

// IsContext 判断是否是context.Context类型
func IsContext(t reflect.Type) bool {
	return t.Implements(ctxType)
}

// IsError 判断是否是error类型
func IsError(t reflect.Type) bool {
	return t.Implements(errType)
}

// Deref is Indirect for reflect.Types
func Deref(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}
