package server

import (
	"github.com/foredata/nova/netx"
)

func newGroup(s *server, prefix string, middlewares []netx.Middleware) netx.Group {
	g := &group{server: s, prefix: prefix, middlewares: middlewares}
	return g
}

type group struct {
	server      *server
	prefix      string
	middlewares []netx.Middleware
}

func (g *group) Group(prefix string, middlewares ...netx.Middleware) netx.Group {
	m := append(g.middlewares, middlewares...)
	return newGroup(g.server, g.prefix+prefix, m)
}

func (g *group) Use(middlewares ...netx.Middleware) {
	g.middlewares = append(g.middlewares, middlewares...)
}

func (s *group) CONNECT(path string, handler interface{}, middlewares ...netx.Middleware) {
	s.server.add(netx.MethodConnect, path, 0, handler, middlewares)
}

func (s *group) DELETE(path string, handler interface{}, middlewares ...netx.Middleware) {
	s.server.add(netx.MethodDelete, path, 0, handler, middlewares)
}

func (s *group) GET(path string, handler interface{}, middlewares ...netx.Middleware) {
	s.server.add(netx.MethodGet, path, 0, handler, middlewares)
}

func (s *group) HEAD(path string, handler interface{}, middlewares ...netx.Middleware) {
	s.server.add(netx.MethodHead, path, 0, handler, middlewares)
}

func (s *group) OPTIONS(path string, handler interface{}, middlewares ...netx.Middleware) {
	s.server.add(netx.MethodOptions, path, 0, handler, middlewares)
}

func (s *group) PATCH(path string, handler interface{}, middlewares ...netx.Middleware) {
	s.server.add(netx.MethodPatch, path, 0, handler, middlewares)
}

func (s *group) POST(path string, handler interface{}, middlewares ...netx.Middleware) {
	s.server.add(netx.MethodPost, path, 0, handler, middlewares)
}

func (s *group) PUT(path string, handler interface{}, middlewares ...netx.Middleware) {
	s.server.add(netx.MethodPut, path, 0, handler, middlewares)
}

func (s *group) TRACE(path string, handler interface{}, middlewares ...netx.Middleware) {
	s.server.add(netx.MethodTrace, path, 0, handler, middlewares)
}

func (s *group) Any(path string, handler interface{}, middlewares ...netx.Middleware) {
	s.server.add(netx.MethodAny, path, 0, handler, middlewares)
}
