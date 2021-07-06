package server

import (
	"context"

	"github.com/foredata/nova/netx"
)

type ctxKey struct{}

type scontext struct {
	req       netx.Request
	rspHeader netx.Header
}

func newContext(parent context.Context, sctx *scontext) context.Context {
	return context.WithValue(parent, ctxKey{}, sctx)
}

func getCtx(ctx context.Context) *scontext {
	sctx, _ := ctx.Value(ctxKey{}).(*scontext)
	return sctx
}

// GetRequest 从Context中获取netx.Request
func GetRequest(ctx context.Context) netx.Request {
	sctx := getCtx(ctx)
	if sctx != nil {
		return sctx.req
	}

	return nil
}

// GetResponseHeader 从Context中获取response header
func GetResponseHeader(ctx context.Context) netx.Header {
	sctx := getCtx(ctx)
	if sctx != nil {
		return sctx.rspHeader
	}

	return nil
}

// SetResponseHeader 设置Response header
func SetResponseHeader(ctx context.Context, header netx.Header) {
	sctx := getCtx(ctx)
	if sctx != nil {
		sctx.rspHeader = header
	}
}
