package client

import (
	"sync"

	"github.com/foredata/nova/netx"
)

// Future 用于获取异步结果
type Future interface {
	Add()
	Done(rsp netx.Response, err error)
	Wait() (netx.Response, error)
	Recycle()
}

func NewFuture() Future {
	f := gFuturePool.Get().(*future)
	f.rsp = nil
	f.err = nil
	f.count = 1
	return f
}

var gFuturePool = sync.Pool{
	New: func() interface{} {
		f := &future{}
		f.cond = sync.NewCond(&f.mux)
		return f
	},
}

type future struct {
	cond  *sync.Cond
	mux   sync.Mutex
	count int
	rsp   netx.Response
	err   error
}

func (f *future) Add() {
	f.mux.Lock()
	f.count++
	f.mux.Unlock()
}

func (f *future) Done(rsp netx.Response, err error) {
	f.mux.Lock()
	f.count--
	f.rsp = rsp
	if err != nil {
		f.err = err
	}
	notify := f.count <= 0 || f.err != nil
	f.mux.Unlock()

	if notify {
		f.cond.Signal()
	}
}

func (f *future) Wait() (netx.Response, error) {
	f.mux.Lock()
	for f.count > 0 && f.err == nil {
		f.cond.Wait()
	}
	rsp := f.rsp
	err := f.err
	f.mux.Unlock()

	return rsp, err
}

func (f *future) Recycle() {
	gFuturePool.Put(f)
}
