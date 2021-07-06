package contexts

import (
	"context"
	"sync"
)

// 原生context使用链式查找,这里扩展了context，使用map存储数据
type ctxKey struct{}

func Get(ctx context.Context, key interface{}) interface{} {
	dict, ok := ctx.Value(ctxKey{}).(*sync.Map)
	if !ok {
		return nil
	}

	value, ok := dict.Load(key)
	if !ok {
		return nil
	}
	return value
}

func Set(ctx context.Context, key interface{}, value interface{}) context.Context {
	dict, ok := ctx.Value(ctxKey{}).(*sync.Map)
	if !ok {
		dict = &sync.Map{}
		ctx = context.WithValue(ctx, ctxKey{}, dict)
	}
	dict.Store(key, value)
	return ctx
}

func Del(ctx context.Context, key interface{}) {
	dict, ok := ctx.Value(ctxKey{}).(sync.Map)
	if ok {
		dict.Delete(key)
	}
}
