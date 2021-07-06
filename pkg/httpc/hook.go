package httpc

import (
	"context"
	"net/http"
)

var gNoopHook = &BaseHook{}

// Hook 回调函数,可增加签名等通用处理
type Hook interface {
	OnRequest(ctx context.Context, req *http.Request) error
	OnResponse(ctx context.Context, rsp *http.Response) error
	OnError(ctx context.Context, err error)
}

type BaseHook struct {
}

func (BaseHook) OnRequest(ctx context.Context, req *http.Request) error {
	return nil
}

func (BaseHook) OnResponse(ctx context.Context, rsp *http.Response) error {
	return nil
}

func (BaseHook) OnError(ctx context.Context, err error) {
}
