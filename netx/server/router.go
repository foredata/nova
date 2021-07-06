package server

import (
	"fmt"
	"strings"

	"github.com/foredata/nova/netx"
)

// NewRouter 创建router
func NewRouter() netx.Router {
	r := &router{
		staticMap: make(map[string]*netx.Route),
		nameMap:   make(map[string]*netx.Route),
	}
	return r
}

type router struct {
	noRoute   netx.Callback
	indexed   []*netx.Route          // 通过cmdId索引
	nameMap   map[string]*netx.Route // 通过Name映射
	staticMap map[string]*netx.Route // 通过path静态路由,不存在通配符的情况
	wildcard  tree                   // 通过path动态路由,存在通配符的情况
	routes    []*netx.Route          // 所有routes
}

func (r *router) Find(packet netx.Packet) netx.Callback {
	ident := packet.Identifier()
	// find by cmdId
	if ident.CmdID != 0 && ident.CmdID < uint32(len(r.indexed)) {
		route := r.indexed[ident.CmdID]
		if route != nil {
			return route.Callback
		}
	}

	// RPC不使用method
	if ident.Method == netx.MethodUnknown {
		if route, ok := r.nameMap[ident.URI]; ok {
			return route.Callback
		}

		return r.noRoute
	}

	// find by static path
	path := ""
	url := ident.URL()
	if url != nil {
		path = url.Path
	}

	if path == "" {
		return r.noRoute
	}

	key := toMethodPath(ident.Method, path)
	if route, ok := r.staticMap[key]; ok {
		return route.Callback
	}

	// any method
	key = toMethodPath(netx.MethodAny, path)
	if route, ok := r.staticMap[key]; ok {
		return route.Callback
	}

	// find by dynamic path

	return r.noRoute
}

func (r *router) Routes() []*netx.Route {
	return r.routes
}

func (r *router) NoRoute(callback netx.Callback) {
	r.noRoute = callback
}

func (r *router) Register(route *netx.Route) {
	if route.Name == "" {
		route.Name = funcName(route.Handler)
	}

	if route.CmdID != 0 {
		// add by command ID
		if route.CmdID >= uint(len(r.indexed)) {
			r.indexed = make([]*netx.Route, route.CmdID*2)
		}
		if r.indexed[route.CmdID] != nil {
			panic(fmt.Errorf("duplicate route by cmdId, %+v", route.CmdID))
		}
		r.indexed[route.CmdID] = route
	}

	// rpc use name
	if route.Name != "" {
		r.nameMap[route.Name] = route
	}

	if route.Path == "" {
		return
	}

	if route.Path != "" {
		if !strings.ContainsAny(route.Path, ":*{[(") {
			// static path
			key := toMethodPath(route.Method, route.Path)
			if _, ok := r.staticMap[key]; ok {
				panic(fmt.Errorf("duplicate route,method=%s, path=%+v", route.Method.String(), route.Path))
			}

			r.staticMap[key] = route
		} else {
			// dynamic path
			if err := r.wildcard.Add(route); err != nil {
				panic(err)
			}
		}
	}

	r.routes = append(r.routes, route)
}
