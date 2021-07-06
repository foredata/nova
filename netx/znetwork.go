package netx

import (
	"context"
	"errors"
	"io"
	"net"
)

// network error
var (
	ErrConnOpened          = errors.New("conn is opened.")
	ErrConnClosed          = errors.New("conn is closed")
	ErrFilterIndexOverflow = errors.New("filter index overflow")
	ErrNotSupport          = errors.New("not support")
	ErrClosed              = errors.New("closed")
	ErrInvalidFrame        = errors.New("invalid frame")
	ErrInvalidIdentifier   = errors.New("invalid identifier")
)

// unique key group define
const (
	KeyGroupConn   = "conn"
	KeyGroupFilter = "filter"
)

// Status socket 状态
type Status int

const (
	// CLOSED 关闭状态
	CLOSED = Status(iota)
	// CONNECTING 连接中
	CONNECTING
	// OPEN 已经建立连接
	OPEN
	// CLOSING 关闭中
	CLOSING
)

// Recycler 可回收复用
type Recycler interface {
	Recycle()
}

// WriterTo 用于数据写入socket,每次写入保持原子性,通常分为两种
//	普通消息:消息包一般很小
//	文件类型:通常很大,需要分块写入
type WriterTo interface {
	io.WriterTo
	io.Closer
}

// Conn 异步Socket
type Conn interface {
	ID() uint32                 // 唯一自增ID,由底层自动生成
	Tag() string                // 标签
	Tran() Tran                 // Transport
	Status() Status             // 当前状态
	IsActive() bool             // 是否已经建立好连接
	IsClient() bool             // 是否是通过Dial建立的Client连接
	LocalAddr() string          // 本地地址
	RemoteAddr() string         // 远程地址
	Write(p WriterTo) error     // 异步写数据,线程安全
	Send(msg interface{}) error // 异步发消息,会触发Filter Write操作
	Close() error               // 调用后将不再接收任何读写操作,并等待所有发送完成后再安全关闭
	Attributes() AttributeMap   // 扩展属性
	Protocol() interface{}      // 绑定协议
	SetProtocol(p interface{})  // 设置解析协议
}

// Listener 类似net.Listener,但去除了Accept方法
type Listener interface {
	Close() error
	Addr() net.Addr
}

// Tran 创建Conn,可以是tcp,websocket等协议
// 不同的Tran可以配置不同的FilterChain
type Tran interface {
	String() string
	GetChain() FilterChain
	SetChain(chain FilterChain)
	AddFilters(filters ...Filter)
	Dial(addr string, opts ...Option) (Conn, error)
	Listen(addr string, opts ...Option) (Listener, error)
	Close() error
}

// Factory 用于创建Transport
type Factory func() Tran

// Filter 用于链式处理Conn各种回调
// InBound: 从前向后执行,包括Read,Open,Error
// OutBound:从后向前执行,包括Write,Close
type Filter interface {
	Name() string
	HandleRead(ctx FilterCtx) error
	HandleWrite(ctx FilterCtx) error
	HandleOpen(ctx FilterCtx) error
	HandleClose(ctx FilterCtx) error
	HandleError(ctx FilterCtx) error
}

// FilterCtx Filter上下文，默认会自动调用Next,如需终止，需要主动调用Abort
type FilterCtx interface {
	context.Context           //
	Recycler                  // 可回收复用
	Attributes() AttributeMap // 自定义数据,Ctx运行结束后,则会失效
	Conn() Conn               // Socket Connection
	Data() interface{}        // 获取数据
	SetData(data interface{}) // 设置数据
	Error() error             // 错误信息
	SetError(err error)       // 设置错误信息
	IsAbort() bool            // 是否已经强制终止
	Abort()                   // 终止调用
	Next() error              // 调用下一个
	Jump(index int) error     // 跳转到指定位置,可以是负索引
	JumpBy(name string) error // 通过名字跳转
	Call() error              // 开始执行,执行完成后会释放FilterCtx
	Clone() FilterCtx         // 拷贝当前状态,可用于转移到其他协程中继续执行
}

// FilterChain 管理Filter,并链式调用所有Filter
// Filter分为Inbound和Outbound
// InBound: 从前向后执行,包括Read,Open,Error
// OutBound:从后向前执行,包括Write,Close
type FilterChain interface {
	Len() int                                     // 长度
	Front() Filter                                // 第一个
	Back() Filter                                 // 最后一个
	Get(index int) Filter                         // 通过索引获取filter
	Index(name string) int                        // 通过名字查询索引
	AddFirst(filters ...Filter)                   // 在前边插入
	AddLast(filters ...Filter)                    // 在末尾插入
	HandleOpen(conn Conn)                         // 建立连接
	HandleClose(conn Conn)                        // 关闭连接
	HandleError(conn Conn, err error)             // 发生错误
	HandleRead(conn Conn, msg interface{})        // 读事件
	HandleWrite(conn Conn, msg interface{}) error // 写事件
}
