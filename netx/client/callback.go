package client

import (
	"github.com/foredata/nova/netx"
)

func toCallback(handler interface{}) (netx.Callback, Future) {
	if handler == nil {
		ft := NewFuture()
		cb := func(conn netx.Conn, packet netx.Packet) error {
			rsp := packet.(netx.Response)
			ft.Done(rsp, nil)
			return nil
		}

		return cb, ft
	}

	var callback netx.Callback
	switch h := handler.(type) {
	case func(netx.Response) error:
		callback = func(conn netx.Conn, packet netx.Packet) error {
			rsp := packet.(netx.Response)
			return h(rsp)
		}
	}

	return callback, nil
}
