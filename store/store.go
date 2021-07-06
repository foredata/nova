package store

import (
	"context"
)

// Store kv存储接口以及cas接口
//	后台可基于redis,etcd,consul实现存储
//
// https://github.com/abronan/valkeyrie
// https://etcd.io/
// https://www.consul.io/docs/agent/kv.html
// https://researchlab.github.io/2018/10/07/redis-10-scan/
type Store interface {
	Name() string
	Incr(ctx context.Context, key string, delta int64) error
	Decr(ctx context.Context, key string, delta int64) error
	// Put 更新数据,返回是否更新成功,原数据(需WithGet显示指定)以及错误信息
	Put(ctx context.Context, key string, value []byte, opts ...PutOption) (bool, []byte, error)
	// Get 通过key查询,不存在返回ErrNotFound
	Get(ctx context.Context, key string) ([]byte, error)
	List(ctx context.Context, pattern string) ([]*KVPair, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Scan(ctx context.Context, pattern string, cursor int64, count int) (int64, []*KVPair, error)
	Close() error
}

type KVPair struct {
	Key   string
	Value []byte
}
