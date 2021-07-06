package lru

import (
	"container/list"
	"context"
	"sync"
	"time"

	"github.com/foredata/nova/cache"
	"github.com/foredata/nova/times"
	"github.com/foredata/nova/times/timing"
)

const (
	cacheName = "lru"
)

func init() {
	cache.Add(cacheName, New)
}

// New .
func New(conf *cache.Config) (cache.Backend, error) {
	lc := &lruCache{maxSize: conf.MaxSize}
	return lc, nil
}

// entry is used to hold a value in the evictList
type entry struct {
	key     string      //
	value   interface{} //
	expired time.Time   // 过期时间
	timerID timing.ID   // 定时器ID
}

// IsExpired 判断是否过期
func (e *entry) IsExpired() bool {
	if e.expired.IsZero() {
		return false
	}

	now := times.Now()
	if e.expired.After(now) {
		return false
	}

	return true
}

// lruCache 最简单的LruCache，线程安全
// https://github.com/hashicorp/golang-lru
type lruCache struct {
	mux     sync.RWMutex
	maxSize int
	list    *list.List
	items   map[string]*list.Element
	onEvict cache.EvictCallback
}

func (lc *lruCache) Name() string {
	return cacheName
}

func (lc *lruCache) Len() int {
	lc.mux.RLock()
	size := len(lc.items)
	lc.mux.RUnlock()
	return size
}

func (lc *lruCache) Exists(ctx context.Context, key string) bool {
	lc.mux.RLock()
	_, exists := lc.items[key]
	lc.mux.RUnlock()
	return exists
}

func (lc *lruCache) Get(ctx context.Context, key string) (value interface{}, err error) {
	lc.mux.Lock()

	if elem, ok := lc.items[key]; ok {
		ent := elem.Value.(*entry)
		if ent.IsExpired() {
			lc.removeElement(elem)
			err = cache.ErrNotFound
		} else {
			lc.list.MoveToFront(elem)
			value = ent.value
		}
	}

	lc.mux.Unlock()

	return
}

func (lc *lruCache) Put(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	lc.mux.Lock()
	var ent *entry
	if elem, ok := lc.items[key]; ok {
		lc.list.MoveToFront(elem)
		ent = elem.Value.(*entry)
		ent.value = value
	} else {
		ent = &entry{key: key, value: value}
		elem := lc.list.PushFront(ent)
		lc.items[key] = elem

		if lc.list.Len() > int(lc.maxSize) {
			lc.removeOldest()
		}
	}

	if ttl > 0 {
		timing.Stop(ent.timerID)
		ent.expired = times.Now().Add(ttl)
		ent.timerID = timing.NewTimer(ent.expired, lc.onTimeout, ent)
	}
	lc.mux.Unlock()
	return nil
}

func (lc *lruCache) Delete(ctx context.Context, key string) error {
	lc.mux.Lock()
	if elem, ok := lc.items[key]; ok {
		lc.removeElement(elem)
	}
	lc.mux.Unlock()
	return nil
}

func (lc *lruCache) Clear() {
	lc.mux.Lock()
	for _, elem := range lc.items {
		lc.removeElement(elem)
	}
	lc.mux.Unlock()
}

func (lc *lruCache) onTimeout(data interface{}) {
	ent := data.(*entry)
	lc.mux.Lock()
	if elem, ok := lc.items[ent.key]; ok {
		ent.timerID = 0
		lc.removeElement(elem)
	}
	lc.mux.Unlock()
}

func (lc *lruCache) removeOldest() {
	lc.removeElement(lc.list.Back())
}

func (lc *lruCache) removeElement(e *list.Element) {
	lc.list.Remove(e)
	ent := e.Value.(*entry)
	delete(lc.items, ent.key)
	if lc.onEvict != nil {
		lc.onEvict(ent.key, ent.value)
	}
	timing.Stop(ent.timerID)
}
