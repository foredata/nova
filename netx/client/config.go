package client

import (
	"context"
	"time"

	"github.com/foredata/nova/netx"
)

const (
	defDialTimeout = time.Millisecond * 500
	defCallTimeout = time.Second
)

// Configer 配置参数,可实现此接口从配置控制面获取相关配置
type Configer interface {
	GetDialTimeout(ctx context.Context, req netx.Request) time.Duration
	GetCallTimeout(ctx context.Context, req netx.Request) time.Duration
}

var gDefaultConfig = &defaultConfig{}

type defaultConfig struct {
}

func (c defaultConfig) GetDialTimeout(ctx context.Context, req netx.Request) time.Duration {
	return defDialTimeout
}

func (c defaultConfig) GetCallTimeout(ctx context.Context, req netx.Request) time.Duration {
	return defCallTimeout
}
