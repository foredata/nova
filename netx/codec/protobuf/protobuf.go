package protobuf

import (
	"errors"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/pkg/bytex"
)

var ErrNotMessage = errors.New("not pb message")

// New .
func New() netx.Codec {
	return &protobufCodec{}
}

type protobufCodec struct {
}

func (c *protobufCodec) Type() netx.CodecType {
	return netx.CodecTypeProtobuf
}

func (c *protobufCodec) Name() string {
	return "protobuf"
}

func (c *protobufCodec) Encode(b bytex.Buffer, msg interface{}) error {
	if m, ok := msg.(Message); ok {
		if data, err := Marshal(m); err == nil {
			return b.Append(data)
		} else {
			return err
		}
	}

	return ErrNotMessage
}

func (c *protobufCodec) Decode(b bytex.Buffer, msg interface{}) error {
	if pb, ok := msg.(Message); ok {
		data := b.Bytes()
		return Unmarshal(data, pb)
	}

	return ErrNotMessage
}
