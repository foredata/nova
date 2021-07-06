package xml

import (
	"encoding/xml"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/pkg/bytex"
)

// New .
func New() netx.Codec {
	return &xmlCodec{}
}

type xmlCodec struct {
}

func (c *xmlCodec) Type() netx.CodecType {
	return netx.CodecTypeXml
}

func (c *xmlCodec) Name() string {
	return "xml"
}

func (c *xmlCodec) Encode(b bytex.Buffer, msg interface{}) error {
	enc := xml.NewEncoder(b)
	return enc.Encode(msg)
}

func (c *xmlCodec) Decode(b bytex.Buffer, msg interface{}) error {
	dec := xml.NewDecoder(b)
	return dec.Decode(msg)
}
