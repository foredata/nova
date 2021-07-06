package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/netx/registry"

	// 强制注册codec
	_ "github.com/foredata/nova/netx/codec"
)

// New 创建Server
func New(opts ...Option) netx.Server {
	o := newOptions(opts...)
	s := &server{opts: o}
	return s
}

type server struct {
	opts        *Options
	middlewares []netx.Middleware
	service     *registry.Service
	exit        chan os.Signal
	addr        net.Addr
}

func (s *server) Addr() net.Addr {
	return s.addr
}

func (s *server) Options() *Options {
	return s.opts
}

func (s *server) Run() error {
	if err := s.Start(); err != nil {
		return err
	}

	s.Wait()
	return s.Stop()
}

func (s *server) buildService() error {
	// build service
	if s.opts.Node == nil {
		s.opts.Node = &registry.Node{}
	}

	node := s.opts.Node
	if node.ID == "" {
		node.ID = s.opts.ID
	}
	node.Addr = s.opts.Addr

	service := &registry.Service{
		Name:     s.opts.Name,
		Version:  s.opts.Version,
		Metadata: s.opts.Metadata,
		Nodes:    []*registry.Node{node},
	}

	routes := s.opts.Router.Routes()
	for _, r := range routes {
		ep := &registry.Endpoint{
			Name:     r.Name,
			Metadata: r.Metadata,
		}

		if r.CmdID != 0 {
			ep.Metadata["cmd_id"] = strconv.Itoa(int(r.CmdID))
		}

		if r.Path != "" {
			ep.Metadata["path"] = r.Path
		}

		service.Endpoints = append(service.Endpoints, ep)
	}

	s.service = service

	return nil
}

func (s *server) Start() error {
	opts := s.opts
	l, err := opts.Tran.Listen(opts.Addr)
	if err != nil {
		return err
	}

	s.addr = l.Addr()
	for _, m := range opts.Modules {
		if err := m.Start(); err != nil {
			return fmt.Errorf("[%s] module start fail, %w", m.Name(), err)
		}
	}

	if err := s.buildService(); err != nil {
		return err
	}

	if s.opts.Registry != nil {
		if err := s.opts.Registry.Register(context.Background(), s.service, s.opts.RegistryTTL); err != nil {
			return fmt.Errorf("registry fail, %+v", s.service)
		}

		s.tickRegistry()
	}

	for _, m := range s.opts.Modules {
		if err := m.Start(); err != nil {
			opts.Tran.Close()
			return err
		}
	}

	return nil
}

func (s *server) Stop() error {
	var errList []string

	if s.opts.Registry != nil {
		if err := s.opts.Registry.Deregister(context.Background(), s.service); err != nil {
			errList = append(errList, fmt.Sprintf("deregister fail, %+v", err.Error()))
		}
	}

	if err := s.opts.Tran.Close(); err != nil {
		errList = append(errList, fmt.Sprintf("tran close fail, %+v", err.Error()))
	}

	for _, m := range s.opts.Modules {
		if err := m.Stop(); err != nil {
			errList = append(errList, fmt.Sprintf("[%s] module stop fail, %s", m.Name(), err.Error()))
		}
	}

	if len(errList) > 0 {
		return fmt.Errorf("server stop has some error, %s:\n", strings.Join(errList, "\n"))
	}

	return nil
}

// 定时自动服务注册
func (s *server) tickRegistry() {
	ttl := s.opts.RegistryTTL
	t := time.NewTicker(ttl / 3)
	go func() {
		for range t.C {
			_ = s.opts.Registry.Register(context.Background(), s.service, ttl)
		}
	}()
}

func (s *server) Wait() {
	s.exit = make(chan os.Signal, 1)
	signal.Notify(s.exit, s.opts.Signals...)
	<-s.exit
	s.exit = nil
}

func (s *server) Exit() {
	if s.exit != nil {
		s.exit <- syscall.SIGQUIT
	}
}

func (s *server) Group(prefix string, middlewares ...netx.Middleware) netx.Group {
	return newGroup(s, prefix, middlewares)
}

func (s *server) Use(middlewares ...netx.Middleware) {
	s.middlewares = append(s.middlewares, middlewares...)
}

func (s *server) CONNECT(path string, handler interface{}, middlewares ...netx.Middleware) {
	s.add(netx.MethodConnect, path, 0, handler, middlewares)
}

func (s *server) DELETE(path string, handler interface{}, middlewares ...netx.Middleware) {
	s.add(netx.MethodDelete, path, 0, handler, middlewares)
}

func (s *server) GET(path string, handler interface{}, middlewares ...netx.Middleware) {
	s.add(netx.MethodGet, path, 0, handler, middlewares)
}

func (s *server) HEAD(path string, handler interface{}, middlewares ...netx.Middleware) {
	s.add(netx.MethodHead, path, 0, handler, middlewares)
}

func (s *server) OPTIONS(path string, handler interface{}, middlewares ...netx.Middleware) {
	s.add(netx.MethodOptions, path, 0, handler, middlewares)
}

func (s *server) PATCH(path string, handler interface{}, middlewares ...netx.Middleware) {
	s.add(netx.MethodPatch, path, 0, handler, middlewares)
}

func (s *server) POST(path string, handler interface{}, middlewares ...netx.Middleware) {
	s.add(netx.MethodPost, path, 0, handler, middlewares)
}

func (s *server) PUT(path string, handler interface{}, middlewares ...netx.Middleware) {
	s.add(netx.MethodPut, path, 0, handler, middlewares)
}

func (s *server) TRACE(path string, handler interface{}, middlewares ...netx.Middleware) {
	s.add(netx.MethodTrace, path, 0, handler, middlewares)
}

func (s *server) Any(path string, handler interface{}, middlewares ...netx.Middleware) {
	s.add(netx.MethodAny, path, 0, handler, middlewares)
}

func (s *server) Register(route *netx.Route) {
	s.opts.Router.Register(route)
}

func (s *server) NoRoute(handler interface{}, middlewares ...netx.Middleware) {
	middlewares = append(s.middlewares, middlewares...)
	endpoint := netx.Apply(toEndpoint(handler, s.opts), middlewares)
	callback := toCallback(endpoint)
	s.opts.Router.NoRoute(callback)
}

func (s *server) add(method netx.Method, path string, cmdId uint, handler interface{}, middlewares []netx.Middleware) {
	middlewares = append(s.middlewares, middlewares...)
	endpoint := netx.Apply(toEndpoint(handler, s.opts), middlewares)
	callback := toCallback(endpoint)

	route := &netx.Route{
		Method:      method,
		Path:        path,
		CmdID:       cmdId,
		Handler:     handler,
		Middlewares: middlewares,
		Callback:    callback,
	}

	s.opts.Router.Register(route)
}
