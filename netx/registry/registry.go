package registry

import (
	"context"
	"time"
)

// Registry 服务注册和查询接口
type Registry interface {
	Name() string
	// Register 注册服务,仅注册一次,上层逻辑会自动补偿注册
	Register(ctx context.Context, service *Service, ttl time.Duration) error
	// Deregister 注销服务
	Deregister(ctx context.Context, service *Service) error
	// Get 查询某个服务,可能会返回多个版本
	Get(ctx context.Context, service string) ([]*Service, error)
	// List 获取所有服务
	List(ctx context.Context) ([]*Service, error)
	// Watch 监听服务变化
	Watch(ctx context.Context) (Watcher, error)
	// Close ...
	Close() error
}

// Service 服务相关信息
//	Name和Version唯一确定一个服务
// 	相同的服务应该有相同的Endpoint,metadata信息
// 只有nodes中信息
type Service struct {
	Name      string            `json:"name"`      // 服务名
	Version   string            `json:"version"`   // 版本
	Metadata  map[string]string `json:"metadata"`  // 元信息
	Endpoints []*Endpoint       `json:"endpoints"` // 端点信息
	Nodes     []*Node           `json:"nodes"`     // 节点信息
}

// Node 服务节点,id,addr必须填写,
// tags和weight可以用于服务发现筛选
// 常见的tags有,cluster,env
type Node struct {
	ID       string            `json:"id"`       // 唯一id
	Addr     string            `json:"addr"`     // ip地址和端口
	Metadata map[string]string `json:"metadata"` // 元信息,不可筛选
}

// Endpoint 端点信息,用于描述每一个api接口
//	对于服务发现而言,通常不是必须提供
type Endpoint struct {
	Name     string            `json:"name"`
	Request  *Value            `json:"request"`
	Response *Value            `json:"response"`
	Metadata map[string]string `json:"metadata"`
}

// Value ...
type Value struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Values []*Value `json:"values"`
}
