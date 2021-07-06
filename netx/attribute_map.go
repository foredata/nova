package netx

import (
	"sort"
	"sync"

	"github.com/foredata/nova/pkg/unique"
)

// AttributeMap 存储key-value数据,线程安全
// 	数据通常不多,底层使用有序数据存储
type AttributeMap interface {
	// 属性数目
	Len() int
	// 是否包含key
	Contains(key unique.Key) bool
	// 获取属性,若不存在且creator不为nil,则会创建
	Get(key unique.Key, creator func() interface{}) interface{}
	// 设置属性
	Put(key unique.Key, value interface{})
	// 删除属性
	Remove(key unique.Key) interface{}
	// 清空属性
	Clear()
}

func NewAttributeMap() AttributeMap {
	return &attributeMap{}
}

type attribute struct {
	key   unique.Key
	value interface{}
}

func (a *attribute) Equals(key unique.Key) bool {
	return a.key.ID() == key.ID()
}

type attributeMap struct {
	sync.RWMutex
	attrs []*attribute
}

func (am *attributeMap) Len() int {
	am.RLock()
	defer am.RUnlock()
	return len(am.attrs)
}

func (am *attributeMap) Contains(key unique.Key) bool {
	am.RLock()
	defer am.RUnlock()
	for _, attr := range am.attrs {
		if attr.Equals(key) {
			return true
		}
	}
	return false
}

func (am *attributeMap) Get(key unique.Key, creator func() interface{}) interface{} {
	am.RLock()
	value := am.getValue(key)
	if value != nil {
		am.RUnlock()
		return value
	}
	am.RUnlock()

	if creator != nil {
		am.Lock()
		// double check
		value := am.getValue(key)
		if value != nil {
			am.Unlock()
			return value
		}

		value = creator()
		am.putValue(key, value)
		am.Unlock()
		return value
	}

	return nil
}

func (am *attributeMap) getValue(key unique.Key) interface{} {
	idx := am.indexOf(key)
	if idx != -1 {
		return am.attrs[idx].value
	}

	return nil
}

func (am *attributeMap) putValue(key unique.Key, value interface{}) {
	attr := &attribute{key: key, value: value}
	if len(am.attrs) == 0 {
		am.attrs = append(am.attrs, attr)
		return
	}

	idx := sort.Search(len(am.attrs), func(i int) bool {
		return am.attrs[i].key.ID() <= key.ID()
	})

	if idx == len(am.attrs) {
		am.attrs = append(am.attrs, attr)
		return
	}
	if am.attrs[idx].Equals(key) {
		am.attrs[idx].value = value
		return
	}
	// create
	am.attrs = append(am.attrs, attr)
	copy(am.attrs[idx+1:], am.attrs[idx:])
	am.attrs[idx] = attr
}

func (am *attributeMap) Put(key unique.Key, value interface{}) {
	am.Lock()
	am.putValue(key, value)
	am.Unlock()
}

func (am *attributeMap) Remove(key unique.Key) interface{} {
	am.Lock()
	for idx, attr := range am.attrs {
		if attr.key.ID() == key.ID() {
			am.attrs = append(am.attrs[:idx], am.attrs[idx+1:]...)
			break
		}
	}
	am.Unlock()
	return nil
}

func (am *attributeMap) Clear() {
	am.Lock()
	am.attrs = am.attrs[:]
	am.Unlock()
}

func (am *attributeMap) indexOf(key unique.Key) int {
	idx := sort.Search(len(am.attrs), func(i int) bool {
		return am.attrs[i].key.ID() <= key.ID()
	})

	if idx == len(am.attrs) || am.attrs[idx].key.ID() != key.ID() {
		return -1
	}

	return idx
}
