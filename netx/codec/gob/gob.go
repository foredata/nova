package gob

import (
	"encoding/gob"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/pkg/bytex"
)

// New .
func New() netx.Codec {
	return &gobCodec{}
}

type gobCodec struct {
}

func (c *gobCodec) Type() netx.CodecType {
	return netx.CodecTypeGob
}

func (c *gobCodec) Name() string {
	return "gob"
}

func (c *gobCodec) Encode(b bytex.Buffer, msg interface{}) error {
	enc := gob.NewEncoder(b)
	return enc.Encode(msg)
}

func (c *gobCodec) Decode(b bytex.Buffer, msg interface{}) error {
	dec := gob.NewDecoder(b)
	return dec.Decode(msg)
}
