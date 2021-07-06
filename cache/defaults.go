package cache

import (
	"context"
	"time"
)

var gDefault Cache

func SetDefault(d Cache) {
	gDefault = d
}

// Len 默认Cache Len
func Len() int {
	return gDefault.Len()
}

// Get 使用默认Cache Get
func Get(ctx context.Context, key string, opts ...GetOption) (interface{}, error) {
	return gDefault.Get(ctx, key, opts...)
}

// Put 默认Cache Put
func Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return gDefault.Put(ctx, key, value, ttl)
}

// Delete 默认Cache Delete
func Delete(ctx context.Context, key string) error {
	return gDefault.Delete(ctx, key)
}

// Clear 默认Cache Clear
func Clear() {
	gDefault.Clear()
}
