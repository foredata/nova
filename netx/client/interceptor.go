package client

import "github.com/foredata/nova/netx"

type Interceptor interface {
	OnRequest(netx.Request) error
	OnResponse(netx.Response) error
	OnError(err error)
}
