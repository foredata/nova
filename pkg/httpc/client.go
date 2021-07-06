package httpc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"path"
	"strings"
	"time"
)

// some erorrs
var (
	ErrNoData      = errors.New("httpc: no data")
	ErrNotSupport  = errors.New("httpc: not support")
	ErrInvalidType = errors.New("httpc: invalid type")
)

// Config .
type Config struct {
	Timeout          time.Duration
	DialTimeout      time.Duration
	HandshakeTimeout time.Duration
	KeepAlive        time.Duration
	BaseUrl          string
}

var defaultConfig = &Config{
	Timeout:          time.Second * 60,
	DialTimeout:      time.Second * 60,
	HandshakeTimeout: time.Second * 60,
	KeepAlive:        time.Second * 60,
}

// NewClient 新建client
func NewClient(o *Config) *Client {
	if o == nil {
		o = defaultConfig
	}
	client := &http.Client{
		Timeout: o.Timeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   o.DialTimeout,
				KeepAlive: o.KeepAlive,
			}).DialContext,
			TLSHandshakeTimeout: o.HandshakeTimeout,
		},
	}

	return &Client{client: client, baseUrl: o.BaseUrl}
}

// Client 简单封装http client,方便调用,注意,CallOptions最多只能传1个
//	1: 构建Request
//	2: body编解码
//	3: retry
//	4: hook
//	5: rsp wrapper
type Client struct {
	client  *http.Client
	baseUrl string
}

func (c *Client) Get(ctx context.Context, url string, opts ...*CallOptions) *Response {
	return c.doRequest(ctx, http.MethodGet, url, nil, toCallOptions(opts...))
}

func (c *Client) Head(ctx context.Context, url string, opts ...*CallOptions) *Response {
	return c.doRequest(ctx, http.MethodHead, url, url, toCallOptions(opts...))
}

func (c *Client) Post(ctx context.Context, url string, reqBody interface{}, opts ...*CallOptions) *Response {
	return c.doRequest(ctx, http.MethodPost, url, reqBody, toCallOptions(opts...))
}

func (c *Client) Put(ctx context.Context, url string, reqBody interface{}, opts ...*CallOptions) *Response {
	return c.doRequest(ctx, http.MethodPut, url, reqBody, toCallOptions(opts...))
}

func (c *Client) Patch(ctx context.Context, url string, reqBody interface{}, opts ...*CallOptions) *Response {
	return c.doRequest(ctx, http.MethodPatch, url, reqBody, toCallOptions(opts...))
}

func (c *Client) Delete(ctx context.Context, url string, opts ...*CallOptions) *Response {
	return c.doRequest(ctx, http.MethodDelete, url, nil, toCallOptions(opts...))
}

func (c *Client) Connect(ctx context.Context, url string, opts ...*CallOptions) *Response {
	return c.doRequest(ctx, http.MethodConnect, url, nil, toCallOptions(opts...))
}

func (c *Client) Options(ctx context.Context, url string, opts ...*CallOptions) *Response {
	return c.doRequest(ctx, http.MethodOptions, url, nil, toCallOptions(opts...))
}

func (c *Client) Trace(ctx context.Context, url string, opts ...*CallOptions) *Response {
	return c.doRequest(ctx, http.MethodTrace, url, nil, toCallOptions(opts...))
}

func (c *Client) doRequest(ctx context.Context, method string, url string, reqBody interface{}, opts *CallOptions) *Response {
	hook := opts.Hook
	if hook == nil {
		hook = gNoopHook
	}

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return newError(err)
	}

	if !strings.HasPrefix(url, "http") && strings.Index(url, "://") == -1 {
		baseUrl := opts.BaseUrl
		if baseUrl == "" {
			baseUrl = c.baseUrl
		}
		if baseUrl != "" {
			url = path.Join(baseUrl, url)
		}
	}

	// build path
	url = replacePath(url, opts.Params)

	body, err := encode(opts.ContentType, reqBody)
	if err != nil {
		return newError(err)
	}

	req.Header = opts.Header
	// build query
	if len(opts.Query) > 0 {
		query := req.URL.Query()
		for k, v := range opts.Query {
			for _, x := range v {
				query.Add(k, x)
			}
		}
		req.URL.RawQuery = query.Encode()
	}

	// build cookies
	for _, cookie := range opts.Cookies {
		req.AddCookie(cookie)
	}

	//
	if err := hook.OnRequest(ctx, req); err != nil {
		return newError(err)
	}

	// call with retry
	for i := 0; ; i++ {
		if body != nil {
			req.Body = ioutil.NopCloser(bytes.NewReader(body))
		}
		rsp, err := c.client.Do(req)
		if err == nil {
			if err := hook.OnResponse(ctx, rsp); err != nil {
				return newError(err)
			}

			if rsp.StatusCode != http.StatusOK {
				err = &StatusErr{Code: rsp.StatusCode, Info: rsp.Status}
			}

			return newResponse(rsp, opts.ContentType, err)
		}

		// 超时重试
		if isTimeoutErr(err) && i < opts.Retry {
			continue
		}

		// 其他错误,直接报错
		hook.OnError(ctx, err)
		return newError(err)
	}
}

// isTimeoutErr 判断是否是超时错误
func isTimeoutErr(err error) bool {
	if err, ok := err.(net.Error); ok && err.Timeout() {
		return true
	}

	return false
}

// StatusErr 当返回码非http.StatusOk时，返回此错误
type StatusErr struct {
	Code int    `json:"code"`
	Info string `json:"info"`
}

func (e *StatusErr) Error() string {
	return fmt.Sprintf("invalid http status,code=%+v, info=%+v", e.Code, e.Info)
}

// IsStatusErr 判断是否是StatusErr错误
func IsStatusErr(e error) bool {
	_, ok := e.(*StatusErr)
	return ok
}
