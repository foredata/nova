package http2

import (
	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/pkg/bytex"
)

func New() netx.Protocol {
	return gHttp2Protocol
}

var gHttp2Protocol = &http2Protocol{}

type http2Protocol struct {
}

func (http2Protocol) Name() string {
	return "http2"
}

func (http2Protocol) Detect(p bytex.Peeker) bool {
	return false
}

func (http2Protocol) Decode(conn netx.Conn, buf bytex.Buffer) (netx.Frame, error) {
	return nil, nil
}

func (http2Protocol) Encode(conn netx.Conn, frame netx.Frame) (bytex.Buffer, error) {
	return nil, nil
}
