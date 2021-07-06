package body

import (
	"io"
	"sync"

	"github.com/foredata/nova/pkg/bytex"
)

// NewBufferBody 创建buffer类型的body
func NewBufferBody(data bytex.Buffer) Body {
	b := gBufferPool.Get().(*bufferBody)
	b.buf = data
	return b
}

var gBufferPool = sync.Pool{
	New: func() interface{} {
		return &bufferBody{}
	},
}

type bufferBody struct {
	noWriter
	buf bytex.Buffer
}

func (b *bufferBody) Close() error {
	if b.buf != nil {
		b.buf.Clear()
		b.buf = nil
	}

	return nil
}

func (b *bufferBody) Read(p []byte) (int, error) {
	return b.buf.Read(p)
}

func (b *bufferBody) ReadFast(blocking bool) (bytex.Buffer, error) {
	if b.buf != nil {
		buf := b.buf
		b.buf = nil
		return buf, io.EOF
	}
	return nil, io.EOF
}

func (b *bufferBody) Buffer() (bytex.Buffer, error) {
	return b.buf, nil
}

func (b *bufferBody) End() bool {
	return b.buf == nil || b.buf.Empty()
}
