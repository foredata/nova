package netx

import (
	"sync"

	"github.com/foredata/nova/pkg/bytex"
)

var gFramePool = sync.Pool{
	New: func() interface{} {
		return &frame{}
	},
}

// NewFrame 创建frame
func NewFrame(ftype FrameType, end bool, streamId uint32, ident *Identifier, header Header, payload bytex.Buffer) Frame {
	f := gFramePool.Get().(*frame)
	f.ftype = ftype
	f.end = end
	f.streamId = streamId
	f.ident = ident
	f.header = header
	f.payload = payload
	return f
}

type frame struct {
	ftype    FrameType
	end      bool
	streamId uint32
	ident    *Identifier
	header   Header
	payload  bytex.Buffer
}

func (f *frame) Type() FrameType {
	return f.ftype
}

func (f *frame) EndFlag() bool {
	return f.end
}

func (f *frame) StreamID() uint32 {
	return f.streamId
}

func (f *frame) SetStreamID(v uint32) {
	f.streamId = v
}

func (f *frame) Identifier() *Identifier {
	return f.ident
}

func (f *frame) SetIdentifier(v *Identifier) {
	f.ident = v
}

func (f *frame) Header() Header {
	return f.header
}

func (f *frame) SetHeader(v Header) {
	f.header = v
}

func (f *frame) Trailer() Header {
	return f.header
}

func (f *frame) SetTrailer(v Header) {
	f.header = v
}

func (f *frame) Payload() bytex.Buffer {
	return f.payload
}

func (f *frame) SetPayload(v bytex.Buffer) {
	f.payload = v
}

func (f *frame) Recycle() {
	gFramePool.Put(f)
}
