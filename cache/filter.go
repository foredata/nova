package cache

import "context"

// Filter 用于过滤不存在的key,防止击穿,通常使用BloomFilter实现
type Filter interface {
	// Exists 判断key是否存在,不存在则返回false,直接忽略
	Exists(ctx context.Context, key string) bool
}

var defaultFilter = &noFilter{}

type noFilter struct {
}

func (*noFilter) Exists(ctx context.Context, key string) bool {
	return true
}
