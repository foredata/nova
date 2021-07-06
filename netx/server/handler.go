package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/netx/body"
	_ "github.com/foredata/nova/netx/codec"
	"github.com/foredata/nova/netx/metadata"
)

// some error
var (
	ErrNotFoundCodec = errors.New("not found codec")
	ErrNoBinder      = errors.New("no binder")
)

// toCallback 将endpoint转换成Callback
func toCallback(endpoint netx.Endpoint) netx.Callback {
	return func(conn netx.Conn, packet netx.Packet) error {
		req, _ := packet.(netx.Request)
		sctx := &scontext{req: req}
		ctx := newContext(context.Background(), sctx)
		if len(req.Header()) > 0 {
			ctx = metadata.NewContext(ctx, req.Header())
		}

		rsp, err := endpoint(ctx, req)
		if req.IsOneway() {
			return err
		}

		if rsp == nil {
			rsp = netx.NewResponse()
		}
		rsp.SetSeqID(req.SeqID())

		if len(sctx.rspHeader) != 0 {
			sctx.rspHeader.Merge(rsp.Header())
			rsp.SetHeader(sctx.rspHeader)
		}

		if rsp.Codec() == 0 {
			rsp.SetCodec(req.Codec())
		}

		if err != nil {
			if nerr, ok := err.(netx.Error); ok {
				rsp.SetStatus(int32(nerr.Code()), nerr.Status())
			} else {
				rsp.SetStatus(http.StatusInternalServerError, err.Error())
			}
		}

		return conn.Send(rsp)
	}
}

// toEndpoint 将interface转换成Endpoint
// 支持以下函数签名:
//	1: netx.Endpoint
//	2: func(context.Context, netx.Request) (netx.Response, error)
// 	3: func(context.Context) error
//	4: func(ctx context.Context) (netx.Response, error)
//	5: func(ctx context.Context, req netx.Request) error
//	6: func(context.Context) (*XResponse, error)
//	7: func(context.Context, *XRequest) error
//	8: func(context.Context, *XRequest) (*XResponse, error)
func toEndpoint(handler interface{}, opts *Options) netx.Endpoint {
	switch h := handler.(type) {
	case netx.Endpoint:
		return h
	case func(context.Context, netx.Request) (netx.Response, error):
		return netx.Endpoint(h)
	case func(context.Context) error:
		// 不需要Request和Response
		return func(ctx context.Context, req netx.Request) (netx.Response, error) {
			return nil, h(ctx)
		}
	case func(ctx context.Context) (netx.Response, error):
		// 不需要Request
		return func(ctx context.Context, req netx.Request) (netx.Response, error) {
			return h(ctx)
		}
	case func(ctx context.Context, req netx.Request) error:
		// 不需要Response
		return func(ctx context.Context, req netx.Request) (netx.Response, error) {
			return nil, h(ctx, req)
		}
	}

	rv := reflect.ValueOf(handler)
	rt := rv.Type()
	if rt.Kind() != reflect.Func {
		panic(fmt.Errorf("convert endpoint fail, type is not function"))
	}

	if rt.NumIn() == 0 || rt.NumIn() > 2 || !isContext(rt.In(0)) {
		panic(fmt.Errorf("convert endpoint fail, input params is invalid"))
	}

	if rt.NumOut() == 0 || rt.NumOut() > 2 || !isError(rt.Out(rt.NumOut()-1)) {
		panic(fmt.Errorf("convert endpoint fail, output params is invalid"))
	}

	// 支持以下几种形式
	// func(context.Context) (*XResponse, error)
	// func(context.Context, *XRequest) error
	// func(context.Context, *XRequest) (*XResponse, error)
	switch {
	case rt.NumIn() == 1 && rt.NumOut() == 2:
		// func(context.Context) (*XResponse, error)
		if !isMessage(rt.Out(0)) {
			panic(fmt.Errorf("convert endpoint fail, output[0] is not message"))
		}
		return func(ctx context.Context, req netx.Request) (netx.Response, error) {
			in := []reflect.Value{reflect.ValueOf(ctx)}
			out := rv.Call(in)
			if !out[1].IsNil() {
				return nil, out[1].Interface().(error)
			}

			if !out[0].IsNil() {
				return encode(ctx, req, out[0].Interface(), opts)
			}
			return nil, nil
		}
	case rt.NumIn() == 2 && rt.NumOut() == 1:
		// func(context.Context, *XRequest) error
		if !isMessage(rt.In(1)) {
			panic(fmt.Errorf("convert endpoint fail, input[1] is not message"))
		}

		return func(ctx context.Context, req netx.Request) (netx.Response, error) {
			msg := reflect.New(rt.In(1).Elem())
			if err := decode(ctx, req, msg.Interface(), opts); err != nil {
				return nil, err
			}

			in := []reflect.Value{reflect.ValueOf(ctx), msg}
			out := rv.Call(in)
			if !out[0].IsNil() {
				return nil, out[0].Interface().(error)
			}

			return nil, nil
		}
	case rt.NumIn() == 2 && rt.NumOut() == 2:
		// func(context.Context, *XRequest) (*XResponse, error)
		if !isMessage(rt.In(1)) || !isMessage(rt.Out(0)) {
			panic(fmt.Errorf("convert endpoint fail, input[1] or output[0] is not message"))
		}
		return func(ctx context.Context, req netx.Request) (netx.Response, error) {
			msg := reflect.New(rt.In(1).Elem())
			if err := decode(ctx, req, msg.Interface(), opts); err != nil {
				return nil, err
			}

			in := []reflect.Value{reflect.ValueOf(ctx), msg}
			out := rv.Call(in)
			if !out[1].IsNil() {
				return nil, out[1].Interface().(error)
			}

			if !out[0].IsNil() {
				return encode(ctx, req, out[0].Interface(), opts)
			}
			return nil, nil
		}
	}

	return nil
}

var (
	ctxType = reflect.TypeOf((*context.Context)(nil)).Elem()
	errType = reflect.TypeOf((*error)(nil)).Elem()
)

func isContext(t reflect.Type) bool {
	return t.Implements(ctxType)
}

func isError(t reflect.Type) bool {
	return t.Implements(errType)
}

// 用于粗略检测函数原型中参数是否是消息类型
// 要求:类型是结构体指针
func isMessage(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}

func decode(ctx context.Context, req netx.Request, msg interface{}, opts *Options) error {
	if req.Codec() == 0 {
		req.SetCodec(uint32(opts.Codec))
	}

	if err := opts.Binder.Bind(ctx, req, msg); err != nil {
		return err
	}

	if opts.Validator != nil {
		return opts.Validator.Validate(ctx, msg)
	}

	return nil
}

// encode 编码格式要求与Request中保持一致,因为没有Response指定Codec
func encode(ctx context.Context, req netx.Request, msg interface{}, opts *Options) (netx.Response, error) {
	if rsp, ok := msg.(netx.Response); ok {
		return rsp, nil
	}

	buf, err := netx.Encode(uint(req.Codec()), msg)
	if err != nil {
		return nil, err
	}

	bod := body.NewBufferBody(buf)
	rsp := netx.NewResponse()
	rsp.SetCodec(req.Codec())
	rsp.SetBody(bod)
	return rsp, nil
}
