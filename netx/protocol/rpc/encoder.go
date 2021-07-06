package rpc

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/pkg/bytex"
)

type encoder struct {
}

func (enc *encoder) Encode(frame netx.Frame, first bool) (bytex.Buffer, error) {
	buf := bytex.NewBuffer()
	// reserved for length + flags
	buf.WriteN(binary.MaxVarintLen32 + 2)

	flags := uint16(0)
	if first {
		// write magic if need
		_ = bytex.WriteUint16BE(buf, magicWord)
		setFlag(&flags, magicMask)
	}

	if frame.EndFlag() {
		setFlag(&flags, frameEndMask)
	}

	set2Bits(&flags, uint16(frame.Type()), frameTypeShift)
	_ = bytex.WriteUvarint64(buf, uint64(frame.StreamID()))

	switch frame.Type() {
	case netx.FrameTypeHeader:
		if err := enc.writeHeaderFrame(buf, frame, flags); err != nil {
			return nil, err
		}
	case netx.FrameTypeData:
		if err := enc.writeDataFrame(buf, frame, flags); err != nil {
			return nil, err
		}
	case netx.FrameTypeTrailer:
		if err := enc.writeTrailerFrame(buf, frame, flags); err != nil {
			return nil, err
		}
	default:
		return nil, netx.ErrInvalidFrame
	}

	return buf, nil
}

func (enc *encoder) writeHeaderFrame(buf bytex.Buffer, frame netx.Frame, flags uint16) error {
	ident := frame.Identifier()
	if ident == nil {
		return netx.ErrInvalidIdentifier
	}

	// version
	if ident.Version > 0 {
		setFlag(&flags, versionMask)
		_ = bytex.WriteUvarint32(buf, uint32(ident.Version))
	}

	// codec+compressType
	// 尚未定义CompressType,仅定义Codec,不超过16
	codecAndCompress := ident.Codec & 0x0F
	_ = bytex.WriteByte(buf, uint8(codecAndCompress))
	_ = bytex.WriteUvarint32(buf, ident.SeqID)

	msgType := ident.MsgType()
	set2Bits(&flags, uint16(msgType-1), msgTypeShift)
	switch msgType {
	case netx.MsgTypeCall, netx.MsgTypeOneway:
		if ident.CmdID > 0 {
			setFlag(&flags, cmdIdMask)
			_ = bytex.WriteUvarint64(buf, uint64(ident.CmdID))
		} else {
			_ = bytex.WriteString(buf, ident.URI)
		}
	case netx.MsgTypeException:
		_ = bytex.WriteVarint32(buf, int32(ident.StatusCode))
		_ = bytex.WriteString(buf, ident.StatusInfo)
	}

	// TODO: common header use for bit

	// write header
	if err := enc.encodeHeader(buf, frame.Header()); err != nil {
		return err
	}

	payload := frame.Payload()
	if payload != nil && !payload.Empty() {
		_ = buf.Append(payload)
	}

	enc.fixLengthFlag(buf, flags)
	return nil
}

// 数据帧,没有额外header,追加数据即可
func (enc *encoder) writeDataFrame(buf bytex.Buffer, frame netx.Frame, flags uint16) error {
	payload := frame.Payload()
	if payload != nil && !payload.Empty() {
		_ = buf.Append(payload)
	}

	enc.fixLengthFlag(buf, flags)
	return nil
}

func (enc *encoder) writeTrailerFrame(buf bytex.Buffer, frame netx.Frame, flags uint16) error {
	// 只会存在一个trailer,end flag一定会设置为true
	setFlag(&flags, frameEndMask)
	if err := enc.encodeHeader(buf, frame.Trailer()); err != nil {
		return err
	}
	enc.fixLengthFlag(buf, flags)
	return nil
}

// encodeHeader 编码header,count + key + values.Join(nullStr)
func (enc *encoder) encodeHeader(buf bytex.Buffer, header netx.Header) error {
	if header.Len() > maxHeaderNum {
		return fmt.Errorf("to header num, %+v", header.Len())
	}
	_ = bytex.WriteUvarint16(buf, uint16(header.Len()))
	header.Walk(func(key string, values []string) bool {
		if len(values) == 0 {
			return true
		}
		val := strings.Join(values, nullStr)
		_ = bytex.WriteString(buf, key)
		_ = bytex.WriteString(buf, val)
		return true
	})

	return nil
}

func (enc *encoder) fixLengthFlag(buf bytex.Buffer, flags uint16) {
	length := uint32(buf.Len() - maxLengthBytes)
	var lengthBytes [5]byte
	n := binary.PutUvarint(lengthBytes[:], uint64(length))

	offset := int64(maxLengthBytes - n)
	_, _ = buf.Seek(offset, io.SeekStart)
	_, _ = buf.Write(lengthBytes[:n])
	_ = bytex.WriteUint16BE(buf, uint16(flags))
	_, _ = buf.Seek(offset, io.SeekStart)
	buf.Discard()
}
