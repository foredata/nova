package store

import (
	"context"
	"time"

	"github.com/foredata/nova/cache"
	"github.com/foredata/nova/encoding"
	"github.com/foredata/nova/store"
)

const (
	cacheName = "store"
)

// some key in config
const (
	StoreKey = "store"
	CodecKey = "codec"
)

func init() {
	cache.Add(cacheName, New)
}

// New .
func New(conf *cache.Config) (cache.Backend, error) {
	store, ok := conf.Get(StoreKey).(store.Store)
	if ok {
		return nil, cache.ErrInvalidConfig
	}

	sc := &storeCache{storage: store, codec: conf.Codec, creator: conf.Creator}
	if sc.codec == nil {
		sc.codec = encoding.NewJsonCodec()
	}

	return sc, nil
}

// storeCache基于store的实现
type storeCache struct {
	storage store.Store
	codec   encoding.Codec
	creator cache.CreatorFunc
}

func (sc *storeCache) Name() string {
	return cacheName
}

func (sc *storeCache) Len() int {
	// not support
	return 0
}

func (sc *storeCache) Exists(ctx context.Context, key string) bool {
	ok, err := sc.storage.Exists(ctx, key)
	return err == nil && ok
}

func (sc *storeCache) Get(ctx context.Context, key string) (interface{}, error) {
	data, err := sc.storage.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if sc.creator == nil {
		return data, nil
	}

	obj := sc.creator()
	if err := sc.codec.Unmarshal(data, obj); err != nil {
		return nil, err
	}

	return obj, nil
}

func (sc *storeCache) Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := sc.codec.Marshal(value)
	if err != nil {
		return err
	}

	_, _, err = sc.storage.Put(ctx, key, data, store.WithTTL(ttl))
	return err
}

func (sc *storeCache) Delete(ctx context.Context, key string) error {
	return sc.storage.Delete(ctx, key)
}

func (sc *storeCache) Clear() {
}
