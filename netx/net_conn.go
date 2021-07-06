package netx

import (
	"sync"
	"sync/atomic"

	"github.com/foredata/nova/pkg/bytex"
)

var (
	gMaxConnID uint32
)

// BaseConn 实现最基础的BaseConn功能
type BaseConn struct {
	sync.Mutex              // 锁,用于控制状态等
	id         uint32       // 唯一ID
	tran       Tran         // netx
	rbuf       bytex.Buffer // 读缓存
	wbuf       NetWriter    // 写缓存
	tag        string       // 标签
	localAddr  string       // 本地地址
	remoteAddr string       // 远程地址
	status     uint32       // 当前状态
	client     bool         // 是否dial产生的连接
	attrs      AttributeMap // kv数据
	protocol   interface{}  // 解析协议
}

// Init ...
func (c *BaseConn) Init(tran Tran, client bool, tag string) {
	id := atomic.AddUint32(&gMaxConnID, 1)
	c.id = id
	c.tran = tran
	c.tag = tag
	c.client = client
	c.rbuf = bytex.NewBuffer()
	c.attrs = NewAttributeMap()
}

func (c *BaseConn) Attributes() AttributeMap {
	return c.attrs
}

// GetChain ...
func (c *BaseConn) GetChain() FilterChain {
	return c.tran.GetChain()
}

// ID ...
func (c *BaseConn) ID() uint32 {
	return c.id
}

// Tag ...
func (c *BaseConn) Tag() string {
	return c.tag
}

// Tran ...
func (c *BaseConn) Tran() Tran {
	return c.tran
}

// Status ...
func (c *BaseConn) Status() Status {
	return Status(atomic.LoadUint32(&c.status))
}

// SetStatus ...
func (c *BaseConn) SetStatus(s Status) {
	atomic.StoreUint32(&c.status, uint32(s))
}

// IsStatus ...
func (c *BaseConn) IsStatus(s Status) bool {
	return c.Status() == s
}

// IsActive ...
func (c *BaseConn) IsActive() bool {
	return c.IsStatus(OPEN)
}

// IsClient ...
func (c *BaseConn) IsClient() bool {
	return c.client
}

// LocalAddr ...
func (c *BaseConn) LocalAddr() string {
	return c.localAddr
}

// SetLocalAddr ...
func (c *BaseConn) SetLocalAddr(addr string) {
	c.localAddr = addr
}

// RemoteAddr ...
func (c *BaseConn) RemoteAddr() string {
	return c.remoteAddr
}

// SetRemoteAddr ...
func (c *BaseConn) SetRemoteAddr(addr string) {
	c.remoteAddr = addr
}

// GetReadBuffer ...
func (c *BaseConn) GetReadBuffer() bytex.Buffer {
	return c.rbuf
}

// GetWriter 获取写缓存
func (c *BaseConn) GetWriter() *NetWriter {
	return &c.wbuf
}

// Write 写缓存
func (c *BaseConn) Write(buff WriterTo) error {
	c.Lock()
	defer c.Unlock()
	if c.IsStatus(CLOSED) || c.IsStatus(CLOSING) {
		return ErrConnClosed
	}
	c.wbuf.Append(buff)
	return nil
}

// Protocol 解析协议
func (c *BaseConn) Protocol() interface{} {
	return c.protocol
}

func (c *BaseConn) SetProtocol(p interface{}) {
	c.protocol = p
}
