package server

import (
	"context"
	"errors"
	"reflect"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/pkg/strx"
)

const (
	tagParam = "param"
	tagQuery = "query"
)

var defaultBinder = &binder{}

// Binder is the interface that wraps the Bind method.
type Binder interface {
	Bind(ctx context.Context, req netx.Request, out interface{}) error
}

// BindUnmarshaler is the interface used to wrap the UnmarshalParam method.
// Types that don't implement this, but do implement encoding.TextUnmarshaler
// will use that interface instead.
type BindUnmarshaler interface {
	// UnmarshalParam decodes and assigns a value from an form or query param.
	UnmarshalParam(param string) error
}

type binder struct {
}

func (b *binder) Bind(ctx context.Context, req netx.Request, out interface{}) error {
	if out == nil {
		return nil
	}

	if err := b.bindPath(ctx, req, out); err != nil {
		return err
	}

	if err := b.bindQuery(ctx, req, out); err != nil {
		return err
	}

	// 不自动绑定header,因为无法快速忽略header,必须执行一次反射

	return b.bindBody(ctx, req, out)
}

func (b *binder) bindPath(ctx context.Context, req netx.Request, out interface{}) error {
	params := req.Params()
	if params.Len() == 0 {
		return nil
	}

	query := map[string][]string{}
	for i, key := range params.Keys() {
		query[key] = []string{params.Values()[i]}
	}

	return bindStruct(out, query, tagParam)
}

func (b *binder) bindQuery(ctx context.Context, req netx.Request, out interface{}) error {
	url := req.URL()
	if url == nil {
		return nil
	}

	query := url.Query()
	if len(query) == 0 {
		return nil
	}

	return bindStruct(out, query, tagQuery)
}

func (b *binder) bindBody(ctx context.Context, req netx.Request, out interface{}) error {
	if req.Body() == nil || req.Body().End() {
		// donot need bind body if no data
		return nil
	}

	// GET, HEAD, DELETE, or OPTIONS  usually don't need one
	buf, _ := req.Body().ReadFast(false)
	return netx.Decode(buf, uint(req.Codec()), out)
}

func bindStruct(dest interface{}, data map[string][]string, tag string) error {
	if dest == nil {
		return nil
	}

	typ := reflect.TypeOf(dest).Elem()
	val := reflect.ValueOf(dest).Elem()
	if typ.Kind() == reflect.Map {
		for k, v := range data {
			val.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v[0]))
		}
		return nil
	}

	if typ.Kind() != reflect.Struct {
		return errors.New("binding element must be a struct")
	}

	for i := 0; i < typ.NumField(); i++ {
		tf := typ.Field(i)
		vf := val.Field(i)
		if !vf.CanSet() {
			continue
		}

		fieldName := tf.Tag.Get(tag)
		if fieldName == "" {
			continue
		}

		values, exists := data[fieldName]
		if !exists {
			continue
		}

		err := strx.BindSlicev(values, vf)
		if err != nil {
			return err
		}
	}

	return nil
}
