// Package metadata is a way of defining message headers
package metadata

import (
	"context"
	"sort"
	"strings"
)

const (
	forwardKey = "RPC_"
)

type metadataKey struct{}

type Pair struct {
	Key    string
	Values []string
}

// Metadata 使用有序数组保存,而不使用map,因为map在序列化时不能保证顺序
type Metadata []Pair

func New() Metadata {
	md := make([]Pair, 0)
	return md
}

func (md Metadata) Len() int {
	return len(md)
}

// Add adds the key, value pair to the header.
// It appends to any existing values associated with key.
func (md *Metadata) Add(key, value string) {
	if key == "" || value == "" {
		return
	}

	kv := md.mustGet(key)
	kv.Values = append(kv.Values, value)
}

// Set sets the header entries associated with key to
// the single element value. It replaces any existing
// values associated with key.
func (md *Metadata) Set(key, value string) {
	if key == "" {
		return
	}
	kv := md.mustGet(key)
	kv.Values = []string{value}
}

func (md *Metadata) SetValues(key string, values []string) {
	if key == "" || len(values) == 0 {
		return
	}

	kv := md.mustGet(key)
	kv.Values = values
}

// Get gets the first value associated with the given key.
func (md Metadata) Get(key string) string {
	idx := md.indexOf(key)
	if idx == -1 {
		return ""
	}
	return md[idx].Values[0]
}

// Values returns all values associated with the given key.
func (md Metadata) Values(key string) []string {
	idx := md.indexOf(key)
	if idx == -1 {
		return nil
	}

	return md[idx].Values
}

// Del deletes the values associated with key.
func (md *Metadata) Del(key string) {
	idx := md.indexOf(key)
	if idx != -1 {
		slice := *md
		copy(slice[idx:], slice[idx+1:])
		*md = slice[:len(slice)-1]
	}
}

// Merge join two metadata
func (md *Metadata) Merge(from Metadata) {
	if len(from) == 0 {
		return
	}
	for _, kv := range from {
		md.SetValues(kv.Key, kv.Values)
	}
}

func (md Metadata) Walk(fn func(key string, values []string) bool) {
	for _, kv := range md {
		if !fn(kv.Key, kv.Values) {
			break
		}
	}
}

func (md Metadata) indexOf(key string) int {
	idx := sort.Search(len(md), func(i int) bool {
		return key <= md[i].Key
	})
	if idx == len(md) || md[idx].Key != key {
		return -1
	}
	return idx
}

func (md *Metadata) mustGet(key string) *Pair {
	slice := *md
	size := len(slice)
	idx := sort.Search(size, func(i int) bool {
		return key <= slice[i].Key
	})

	if idx == size {
		slice = append(slice, Pair{Key: key})
		*md = slice
		return &(slice[size])
	}

	if slice[idx].Key == key {
		return &(slice[idx])
	}

	// create
	slice = append(slice, Pair{})
	copy(slice[idx+1:], slice[idx:])
	slice[idx] = Pair{Key: key}
	*md = slice
	return &(slice[idx])
}

func Add(ctx context.Context, k, v string) context.Context {
	if k == "" || v == "" {
		return ctx
	}

	md, ok := FromContext(ctx)
	if !ok {
		md = New()
		md.Set(k, v)
		return NewContext(ctx, md)
	}

	md.Add(k, v)
	return ctx
}

// Set add key with val to metadata
func Set(ctx context.Context, k, v string) context.Context {
	if k == "" || v == "" {
		return ctx
	}

	md, ok := FromContext(ctx)
	if !ok {
		md = New()
		md.Set(k, v)
		return NewContext(ctx, md)
	}

	md.Set(k, v)
	return ctx
}

// Get returns a single value from metadata in the context
func Get(ctx context.Context, key string) string {
	md, ok := FromContext(ctx)
	if !ok {
		return ""
	}

	return md.Get(key)
}

// Values .
func Values(ctx context.Context, key string) []string {
	md, ok := FromContext(ctx)
	if !ok {
		return nil
	}

	return md.Values(key)
}

// Forward 透传
func Forward(ctx context.Context, old Metadata) Metadata {
	md, ok := FromContext(ctx)
	if !ok {
		return old
	}

	md.Walk(func(key string, values []string) bool {
		if strings.HasPrefix(key, forwardKey) {
			old.SetValues(key, values)
		}

		return true
	})

	return old
}

// NewContext creates a new context with the given metadata
func NewContext(parent context.Context, metadata Metadata) context.Context {
	return context.WithValue(parent, metadataKey{}, metadata)
}

// FromContext returns metadata from the given context
func FromContext(ctx context.Context) (Metadata, bool) {
	md, ok := ctx.Value(metadataKey{}).(Metadata)
	return md, ok
}
