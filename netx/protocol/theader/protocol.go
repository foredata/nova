package theader

import (
	"encoding/binary"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/pkg/bytex"
)

var gTHeaderProtocol = &theaderProtocol{}

// New .
func New() netx.Protocol {
	return gTHeaderProtocol
}

// https://github.com/apache/thrift/blob/master/doc/specs/HeaderFormat.md
// https://github.com/facebook/fbthrift/blob/master/thrift/doc/HeaderFormat.md
// HeaderFormat:
// 0 1 2 3 4 5 6 7 8 9 a b c d e f 0 1 2 3 4 5 6 7 8 9 a b c d e f
// +----------------------------------------------------------------+
// | 0|                          LENGTH                             |
// +----------------------------------------------------------------+
// | 0|       HEADER MAGIC          |            FLAGS              |
// +----------------------------------------------------------------+
// |                         SEQUENCE NUMBER                        |
// +----------------------------------------------------------------+
// | 0|     Header Size(/32)        | ...
// +---------------------------------
//
//                   Header is of variable size:
//                    (and starts at offset 14)
//
// +----------------------------------------------------------------+
// |         PROTOCOL ID  (varint)  |   NUM TRANSFORMS (varint)     |
// +----------------------------------------------------------------+
// |      TRANSFORM 0 ID (varint)   |        TRANSFORM 0 DATA ...
// +----------------------------------------------------------------+
// |         ...                              ...                   |
// +----------------------------------------------------------------+
// |        INFO 0 ID (varint)      |       INFO 0  DATA ...
// +----------------------------------------------------------------+
// |         ...                              ...                   |
// +----------------------------------------------------------------+
// |                                                                |
// |                              PAYLOAD                           |
// |                                                                |
// +----------------------------------------------------------------+
type theaderProtocol struct {
}

func (*theaderProtocol) Name() string {
	return "theader"
}

func (*theaderProtocol) Detect(p bytex.Peeker) bool {
	var data [8]byte
	if n, err := p.Peek(data[:]); err == nil || n != 8 {
		return false
	}

	secondword := binary.BigEndian.Uint32(data[4:])
	return (secondword & magicMask) == headerMagic
}

func (*theaderProtocol) Decode(conn netx.Conn, buf bytex.Buffer) (netx.Frame, error) {
	dec := decoder{}
	return dec.Decode(conn, buf)
}

func (*theaderProtocol) Encode(conn netx.Conn, frame netx.Frame) (bytex.Buffer, error) {
	enc := encoder{}
	return enc.Encode(conn, frame)
}
