package json

import (
	"encoding/json"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/pkg/bytex"
)

// New .
func New() netx.Codec {
	return &jsonCodec{}
}

type jsonCodec struct {
}

func (c *jsonCodec) Type() netx.CodecType {
	return netx.CodecTypeJson
}

func (c *jsonCodec) Name() string {
	return "json"
}

func (c *jsonCodec) Encode(b bytex.Buffer, msg interface{}) error {
	enc := json.NewEncoder(b)
	return enc.Encode(msg)
}

func (c *jsonCodec) Decode(b bytex.Buffer, msg interface{}) error {
	dec := json.NewDecoder(b)
	return dec.Decode(msg)
}
