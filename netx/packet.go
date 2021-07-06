package netx

import (
	"net/url"
	"sync"

	"github.com/foredata/nova/netx/body"
)

func NewRequest() Request {
	return newPacket()
}

func NewResponse() Response {
	rsp := newPacket()
	rsp.ensure()
	rsp.ident.IsResponse = true
	return rsp
}

func NewPacket() Packet {
	return newPacket()
}

var gPacketPool = sync.Pool{
	New: func() interface{} {
		return &packet{}
	},
}

func newPacket() *packet {
	p := gPacketPool.Get().(*packet)
	if p.ident != nil {
		*p.ident = Identifier{}
	}
	return p
}

type packet struct {
	ident   *Identifier
	header  Header
	trailer Header
	body    Body
}

func (p *packet) Recycle() {
	p.ident = nil
	p.header = nil
	p.trailer = nil
	p.body = nil
	gPacketPool.Put(p)
}

func (p *packet) Identifier() *Identifier {
	return p.ident
}

func (p *packet) SetIdentifier(v *Identifier) {
	p.ident = v
}

func (p *packet) Header() Header {
	return p.header
}

func (p *packet) SetHeader(v Header) {
	p.header = v
}

func (p *packet) Trailer() Header {
	return p.trailer
}

func (p *packet) SetTrailer(v Header) {
	p.trailer = v
}

func (p *packet) Body() Body {
	return p.body
}

func (p *packet) SetBody(v Body) {
	p.body = v
}

func (p *packet) Version() uint {
	return p.ident.Version
}

func (p *packet) SetVersion(v uint) {
	p.ensure()
	p.ident.Version = v
}

func (p *packet) SeqID() uint32 {
	return p.ident.SeqID
}

func (p *packet) SetSeqID(v uint32) {
	p.ensure()
	p.ident.SeqID = v
}

func (p *packet) Codec() uint32 {
	return p.ident.Codec
}

func (p *packet) SetCodec(v uint32) {
	p.ensure()
	p.ident.Codec = v
}

func (p *packet) IsOneway() bool {
	return p.ident.IsOneway
}

func (p *packet) SetOneway(v bool) {
	p.ensure()
	p.ident.IsOneway = v
}

func (p *packet) CmdID() uint32 {
	return p.ident.CmdID
}

func (p *packet) SetCmdID(v uint32) {
	p.ensure()
	p.ident.CmdID = v
}

func (p *packet) Service() string {
	return p.ident.Service
}

func (p *packet) SetService(v string) {
	p.ensure()
	p.ident.Service = v
}

func (p *packet) URI() string {
	return p.ident.URI
}

func (p *packet) SetURI(v string) {
	p.ensure()
	p.ident.URI = v
}

func (p *packet) URL() *url.URL {
	return p.ident.url
}

func (p *packet) Params() Params {
	return p.ident.Params
}

func (p *packet) Method() Method {
	return p.ident.Method
}

func (p *packet) SetMethod(v Method) {
	p.ensure()
	p.ident.Method = v
}

func (p *packet) StatusCode() int32 {
	return p.ident.StatusCode
}

func (p *packet) StatusInfo() string {
	return p.ident.StatusInfo
}

func (p *packet) SetStatus(code int32, info string) {
	p.ensure()
	p.ident.StatusCode = code
	p.ident.StatusInfo = info
}

func (p *packet) ensure() {
	if p.ident == nil {
		p.ident = &Identifier{}
	}
}

func (p *packet) Encode(codecType CodecType, msg interface{}) error {
	buf, err := Encode(codecType, msg)
	if err != nil {
		return err
	}
	p.SetCodec(uint32(codecType))
	p.body = body.NewBufferBody(buf)
	return nil
}

func (p *packet) Decode(msg interface{}) error {
	buf, err := p.body.Buffer()
	if err != nil {
		return err
	}

	return Decode(buf, uint(p.ident.Codec), msg)
}
