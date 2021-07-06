package rpc

import (
	"encoding/binary"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/pkg/bytex"
)

var gProtocol = &rpcProtocol{}

func New() netx.Protocol {
	return gProtocol
}

// rpcProtocol 为内部自定义通信协议
// 通信协议定义
// Length[Varient]
// Flags[2Bytes]
// *MagicNumber[2Bytes]
// StreamId[Varient]
// *Version[1Bytes]
// Codec+CompressType[1Byte]
// Request
//	CmdId
//	URI
// Response:
//	StatusCode
//	StatusInfo
// Header
type rpcProtocol struct {
}

func (rp *rpcProtocol) Name() string {
	return "rpc"
}

func (rp *rpcProtocol) Detect(peeker bytex.Peeker) bool {
	// length+flags+magic
	var data [9]byte
	n, _ := peeker.Peek(data[:])
	_, lenSize := binary.Uvarint(data[:n])
	if lenSize <= 0 || lenSize > binary.MaxVarintLen32 {
		return false
	}

	flags := binary.BigEndian.Uint16(data[lenSize:])
	magic := binary.BigEndian.Uint16(data[lenSize+2:])
	if hasFlag(flags, magicMask) && magic == magicWord {
		return true
	}

	return false
}

func (rp *rpcProtocol) Decode(conn netx.Conn, buf bytex.Buffer) (netx.Frame, error) {
	dec := &decoder{}
	return dec.Decode(buf)
}

func (rp *rpcProtocol) Encode(conn netx.Conn, frame netx.Frame) (bytex.Buffer, error) {
	enc := encoder{}
	return enc.Encode(frame, true)
}
