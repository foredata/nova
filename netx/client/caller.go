package client

import (
	"fmt"
	"sync"
	"time"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/times/timing"
)

// Caller 用于跟踪异步回调和超时处理
type Caller interface {
	Register(req netx.Request, callback netx.Callback, timeout time.Duration, retryer Retryer) error
	Unregister(seqId uint32)
	// Find 查询Callback,会自动注销
	Find(packet netx.Packet) netx.Callback
}

func newCaller() Caller {
	c := &caller{}
	c.infos = make(map[uint32]*callInfo)
	return c
}

var gCallPool = sync.Pool{
	New: func() interface{} {
		return &callInfo{}
	},
}

type callInfo struct {
	Request  netx.Request  // 请求消息
	Callback netx.Callback // 回调函数
	Timeout  time.Duration // 超时时间,可以很大但必须指定
	TimerID  uint64        // 定时器唯一ID
	Retryer  Retryer       //
}

func (ci *callInfo) Recyle() {
	ci.Retryer.Recyle()
	gCallPool.Put(ci)
}

func newCallInfo() *callInfo {
	return gCallPool.Get().(*callInfo)
}

type caller struct {
	mux   sync.Mutex
	infos map[uint32]*callInfo
}

func (c *caller) Register(req netx.Request, callback netx.Callback, timeout time.Duration, retryer Retryer) error {
	if req.SeqID() == 0 {
		return fmt.Errorf("zero seqId")
	}

	c.mux.Lock()
	seqId := req.SeqID()
	if _, ok := c.infos[seqId]; ok {
		c.mux.Unlock()
		return fmt.Errorf("duplicate sequence id, %+v", seqId)
	}

	info := newCallInfo()
	info.Request = req
	info.Callback = callback
	info.Timeout = timeout
	info.Retryer = retryer
	info.TimerID = timing.NewDelayer(timeout, c.onTimeout, seqId)
	c.infos[seqId] = info

	c.mux.Unlock()

	return nil
}

func (c *caller) Unregister(seqId uint32) {
	c.mux.Lock()
	info := c.infos[seqId]
	if info != nil {
		timing.Stop(info.TimerID)
		delete(c.infos, seqId)
	}
	c.mux.Unlock()
}

func (c *caller) Find(packet netx.Packet) netx.Callback {
	ident := packet.Identifier()
	if ident == nil || ident.SeqID == 0 {
		return nil
	}

	c.mux.Lock()

	var callback netx.Callback

	info := c.infos[ident.SeqID]
	if info != nil {
		callback = info.Callback
		delete(c.infos, ident.SeqID)
		timing.Stop(info.TimerID)
		info.Recyle()
	}
	c.mux.Unlock()

	return callback
}

// 超时
func (c *caller) onTimeout(data interface{}) {
	seqId := data.(uint32)
	c.mux.Lock()
	info := c.infos[seqId]
	if info == nil {
		c.mux.Unlock()
		return
	}

	delete(c.infos, seqId)

	if !info.Retryer.Allow() {
		callback := info.Callback
		info.Recyle()
		c.mux.Unlock()

		rsp := netx.NewResponse()
		rsp.SetStatus(netx.StatusTimeout, "")
		_ = callback(nil, rsp.(netx.Packet))
		return
	}

	// 重新生成seqId
	seqId = netx.NewSeqID()
	info.Request.SetSeqID(seqId)
	c.infos[seqId] = info
	c.mux.Unlock()

	if err := info.Retryer.Do(info.Request); err != nil {
		c.mux.Lock()
		info.Recyle()
		delete(c.infos, seqId)
		c.mux.Unlock()
	} else {
		timing.NewDelayer(info.Timeout, c.onTimeout, seqId)
	}
}
