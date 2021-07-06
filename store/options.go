package store

import "time"

type PutMode uint8

const (
	PutModeAny PutMode = iota // 不管是否存在
	PutModeXX                 // Only set the key if it already exist.
	PutModeNX                 // Only set the key if it does not already exist.
)

type PutOptions struct {
	Mode PutMode       // 模式
	Get  bool          // 是否返回原数据
	TTL  time.Duration // 过期时间,0不过期
}

type PutOption func(o *PutOptions)

func WithNX() PutOption {
	return func(o *PutOptions) {
		o.Mode = PutModeNX
	}
}

func WithXX() PutOption {
	return func(o *PutOptions) {
		o.Mode = PutModeXX
	}
}

func WithGet() PutOption {
	return func(o *PutOptions) {
		o.Get = true
	}
}

func WithTTL(ttl time.Duration) PutOption {
	return func(o *PutOptions) {
		o.TTL = ttl
	}
}
