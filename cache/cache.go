package cache

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/foredata/nova/errorx"
	"github.com/foredata/nova/pkg/singleflight"
)

// some error
var (
	ErrNotFactory    = errors.New("cache: not find factory")
	ErrInvalidConfig = errors.New("cache: invalid config")
	ErrNotFound      = errors.New("cache: not found")
)

var factories = make(map[string]Factory)

// Add 注册factory
func Add(name string, fn Factory) {
	factories[name] = fn
}

// MustNew 创建cache，异常会抛panic
func MustNew(name string, conf *Config) Cache {
	c, err := New(name, conf)
	if err != nil {
		panic(err)
	}

	return c
}

// New 新建cache
func New(name string, conf *Config) (Cache, error) {
	fact := factories[name]
	if fact == nil {
		return nil, fmt.Errorf("%w, name=%s", ErrNotFactory, name)
	}

	backend, err := fact(conf)
	if err != nil {
		return nil, err
	}
	c := &cache{backend: backend, conf: conf}
	return c, nil
}

// Factory .
type Factory func(conf *Config) (Backend, error)

// Cache 缓存接口,可以是进程内cache,也可以是基于redis的分布式cache
// see: Guava,Caffeine
// 缓存常见问题:穿透,击穿,雪崩
// https://zhuanlan.zhihu.com/p/75588064
//
// 其他库
// https://github.com/dgraph-io/ristretto
//
// TODO: support metrics, statistics, hook
type Cache interface {
	// Len 当前缓存中key的个数,若key已经过期但并没有及时清理，也会计算在内
	Len() int
	// Exists is used to check if the cache contains a key without updating recency or frequency.
	Exists(ctx context.Context, key string) bool
	Get(ctx context.Context, key string, opts ...GetOption) (interface{}, error)
	Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	// Clear 清空所有cache，仅内存cache支持
	Clear()
}

// cache wrapper 用于防止缓存穿透,击穿,雪崩
type cache struct {
	conf    *Config
	backend Backend
	sfg     singleflight.Group
}

func (c *cache) Len() int {
	return c.backend.Len()
}

func (c *cache) Exists(ctx context.Context, key string) bool {
	return c.backend.Exists(ctx, key)
}

func (c *cache) Get(ctx context.Context, key string, opts ...GetOption) (interface{}, error) {
	key = c.conf.Modifier(key)

	if !c.conf.Filter.Exists(ctx, key) {
		return nil, errorx.ErrNotFound
	}

	o := newGetOptions(opts...)

	value, err := c.backend.Get(ctx, key)
	if err == errorx.ErrNotFound {
		if o.Load != nil {
			var v interface{}
			v, err = c.sfg.Do(key, func() (interface{}, error) {
				return o.Load(ctx, key)
			})
			if v != nil {
				value = v.([]byte)
			}
		}

		// 缓存不存在的key,避免恶意攻击
		if err == errorx.ErrNotFound && c.conf.AbsentTTL != 0 {
			_ = c.backend.Put(ctx, key, nil, c.conf.AbsentTTL)
		}
	}

	return value, err
}

func (c *cache) Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// 自动添加抖动
	if ttl != 0 && c.conf.JitterTTL != 0 {
		ttl += time.Duration(rand.Int63n(int64(c.conf.JitterTTL)))
	}

	key = c.conf.Modifier(key)
	return c.backend.Put(ctx, key, value, ttl)
}

func (c *cache) Delete(ctx context.Context, key string) error {
	key = c.conf.Modifier(key)
	return c.backend.Delete(ctx, key)
}

func (c *cache) Clear() {
	c.backend.Clear()
}
