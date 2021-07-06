package cache

import (
	"context"
	"time"
)

// Backend cache 后端实现接口
type Backend interface {
	Name() string
	// Len 当前缓存中key的个数,若key已经过期但并没有及时清理，也会计算在内
	Len() int
	// Exists is used to check if the cache contains a key without updating recency or frequency.
	Exists(ctx context.Context, key string) bool
	Get(ctx context.Context, key string) (interface{}, error)
	// Put ttl为0则不过期
	Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear()
}
