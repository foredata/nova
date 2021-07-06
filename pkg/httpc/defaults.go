package httpc

import (
	"context"
	"net/http"
)

var defaultClient = NewClient(nil)

// SetDefault 设置默认client
func SetDefault(c *Client) {
	defaultClient = c
}

func Get(ctx context.Context, url string, opts ...*CallOptions) *Response {
	return defaultClient.doRequest(ctx, http.MethodGet, url, nil, toCallOptions(opts...))
}

func Head(ctx context.Context, url string, opts ...*CallOptions) *Response {
	return defaultClient.doRequest(ctx, http.MethodHead, url, url, toCallOptions(opts...))
}

func Post(ctx context.Context, url string, reqBody interface{}, opts ...*CallOptions) *Response {
	return defaultClient.doRequest(ctx, http.MethodPost, url, reqBody, toCallOptions(opts...))
}

func Put(ctx context.Context, url string, reqBody interface{}, opts ...*CallOptions) *Response {
	return defaultClient.doRequest(ctx, http.MethodPut, url, reqBody, toCallOptions(opts...))
}

func Patch(ctx context.Context, url string, reqBody interface{}, opts ...*CallOptions) *Response {
	return defaultClient.doRequest(ctx, http.MethodPatch, url, reqBody, toCallOptions(opts...))
}

func Delete(ctx context.Context, url string, opts ...*CallOptions) *Response {
	return defaultClient.doRequest(ctx, http.MethodDelete, url, nil, toCallOptions(opts...))
}

func Connect(ctx context.Context, url string, opts ...*CallOptions) *Response {
	return defaultClient.doRequest(ctx, http.MethodConnect, url, nil, toCallOptions(opts...))
}

func Options(ctx context.Context, url string, opts ...*CallOptions) *Response {
	return defaultClient.doRequest(ctx, http.MethodOptions, url, nil, toCallOptions(opts...))
}

func Trace(ctx context.Context, url string, opts ...*CallOptions) *Response {
	return defaultClient.doRequest(ctx, http.MethodTrace, url, nil, toCallOptions(opts...))
}
