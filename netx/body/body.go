package body

import (
	"errors"
	"io"

	"github.com/foredata/nova/pkg/bytex"
)

// some error
var (
	ErrNotSupport = errors.New("body: not support")
	ErrClosed     = errors.New("body: closed")
)

// New 根据需要创建不同类型的body,通常都为BufferBody
func New(buf bytex.Buffer, streaming bool) Body {
	if streaming {
		return NewStreamBody(buf)
	} else if buf != nil && !buf.Empty() {
		return NewBufferBody(buf)
	} else {
		return gNoBody
	}
}

// Body 消息体接口,应用层只需要读操作
//	支持标准的io.Read接口,效率比较低,会额外拷贝一次内存,当数据不足时,会阻塞
//	ReadFast扩展了Read接口,直接返回底层的Buffer结构,更加高效,支持阻塞和非阻塞两种模式
//	当Read/ReadFast返回io.EOF时,表名没有数据了,但本次读取返回结果中,可能有数据,也可能没数据
type Body interface {
	io.ReadCloser
	// 转换为Buffer,仅支持BufferBody,其他类型会返回NotSupport,用于rpc消息解码
	Buffer() (bytex.Buffer, error)
	// 快速读取,直接返回buffer,返回io.EOF表明没有数据了
	ReadFast(blocking bool) (bytex.Buffer, error)
	// 数据是否已读取完
	End() bool
}

// Writer body数据写入接口,用于streaming模式下,拼接多个frame组成一个完整body
//	底层需要保证读写线程安全
type Writer interface {
	// 写入数据
	Write(data bytex.Buffer) error
	// 调用此接口标记数据传输完成
	Flush()
}
