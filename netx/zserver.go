package netx

import "net"

// Module 用于扩展server
type Module interface {
	Name() string
	Start() error
	Stop() error
}

// Group .
type Group interface {
	Group(prefix string, middlewares ...Middleware) Group
	Use(middlewares ...Middleware)

	CONNECT(path string, handler interface{}, middlewares ...Middleware)
	DELETE(path string, handler interface{}, middlewares ...Middleware)
	GET(path string, handler interface{}, middlewares ...Middleware)
	HEAD(path string, handler interface{}, middlewares ...Middleware)
	OPTIONS(path string, handler interface{}, middlewares ...Middleware)
	PATCH(path string, handler interface{}, middlewares ...Middleware)
	POST(path string, handler interface{}, middlewares ...Middleware)
	PUT(path string, handler interface{}, middlewares ...Middleware)
	TRACE(path string, handler interface{}, middlewares ...Middleware)
}

// Server 服务端接口
type Server interface {
	Group
	Addr() net.Addr        // 服务器监听地址
	Register(route *Route) // 注册router
	NoRoute(handler interface{}, middlewares ...Middleware)
	Run() error
	Exit()
}

// Route 路由信息
type Route struct {
	Name        string            //
	Method      Method            //
	Path        string            //
	CmdID       uint              // 不宜过大尽量保持在uint16以内,底层数组存储
	Metadata    map[string]string // 自定义字段,可用于服务发现中注册额外字段
	Handler     interface{}       // 原始Handler,see handler.go中toEndpoint原型定义
	Middlewares []Middleware      // 中间件
	Callback    Callback          // Handler经过middleware加工后,转换成callback
}

// Router .
type Router interface {
	Routes() []*Route
	Register(route *Route)
	NoRoute(callback Callback)
	Find(packet Packet) Callback
}
