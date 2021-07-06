package rpc

import (
	"fmt"
	"strings"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/pkg/bytex"
)

type decoder struct {
}

func (dec *decoder) Decode(buf bytex.Buffer) (netx.Frame, error) {
	var length uint32
	if err := bytex.ReadUvarint32(buf, &length); err != nil {
		return nil, err
	}
	// 数据不足
	if buf.Available() < int(length) {
		return nil, nil
	}

	realBuf := buf.ReadN(int(length))
	var flags uint16
	if err := bytex.ReadUint16BE(realBuf, &flags); err != nil {
		return nil, err
	}

	if hasFlag(flags, magicMask) {
		var magic uint16
		if err := bytex.ReadUint16BE(realBuf, &magic); err != nil {
			return nil, fmt.Errorf("read magic word fail, %+v", err)
		}
		if magic != magicWord {
			return nil, fmt.Errorf("invalid magic word")
		}
	}

	var streamID uint32
	if err := bytex.ReadUvarint32(realBuf, &streamID); err != nil {
		return nil, fmt.Errorf("read stream id fail, %+v", err)
	}

	frameType := get2Bits(flags, frameTypeShift)
	end := hasFlag(flags, frameEndMask)
	if frameType == netx.FrameTypeTrailer {
		end = true
	}
	frame := netx.NewFrame(netx.FrameType(frameType), end, streamID, nil, nil, nil)
	var err error
	switch frameType {
	case netx.FrameTypeHeader:
		err = dec.readHeaderFrame(frame, realBuf, flags)
	case netx.FrameTypeData:
		err = dec.readDataFrame(frame, realBuf, flags)
	case netx.FrameTypeTrailer:
		err = dec.readTrailerFrame(frame, realBuf, flags)
	default:
		err = netx.ErrNotSupport
	}

	if err != nil {
		frame.Recycle()
		return nil, err
	}

	return frame, nil
}

func (dec *decoder) readHeaderFrame(frame netx.Frame, buf bytex.Buffer, flags uint16) error {
	ident := netx.NewIdentifier()

	// version
	if hasFlag(flags, versionMask) {
		var version uint32
		if err := bytex.ReadUvarint32(buf, &version); err != nil {
			return fmt.Errorf("read version fail, %+v", err)
		}
		ident.Version = uint(version)
	}

	//
	codecAndCompress, err := buf.ReadByte()
	if err != nil {
		return fmt.Errorf("read codec fail, %+v", err)
	}
	ident.Codec = uint32(codecAndCompress & 0x0F)
	if err := bytex.ReadUvarint32(buf, &ident.SeqID); err != nil {
		return fmt.Errorf("read seqid fail, %+v", err)
	}

	msgType := netx.MsgType(get2Bits(flags, msgTypeShift) + 1)
	switch msgType {
	case netx.MsgTypeCall, netx.MsgTypeOneway:
		ident.IsResponse = false
		ident.IsOneway = msgType == netx.MsgTypeOneway
		if hasFlag(flags, cmdIdMask) {
			if err := bytex.ReadUvarint32(buf, &ident.CmdID); err != nil {
				return fmt.Errorf("read cmd id fail, %+v", err)
			}
		} else {
			if err := bytex.ReadString(buf, &ident.URI); err != nil {
				return fmt.Errorf("read uri fail, %+v", err)
			}
		}
	case netx.MsgTypeReply:
		ident.IsResponse = true
	case netx.MsgTypeException:
		ident.IsResponse = true
		if err := bytex.ReadVarint32(buf, &ident.StatusCode); err != nil {
			return fmt.Errorf("read status code fail, %+v", err)
		}
		if err := bytex.ReadString(buf, &ident.StatusInfo); err != nil {
			return fmt.Errorf("read status info fail, %+v", err)
		}
	}
	// read header
	header, err := dec.decodeHeader(buf)
	if err != nil {
		return err
	}
	frame.SetIdentifier(ident)
	frame.SetHeader(header)
	frame.SetPayload(getPayload(buf))
	return nil
}

func (dec *decoder) readDataFrame(frame netx.Frame, buf bytex.Buffer, flags uint16) error {
	frame.SetPayload(getPayload(buf))
	return nil
}

func (dec *decoder) readTrailerFrame(frame netx.Frame, buf bytex.Buffer, flags uint16) error {
	trailer, err := dec.decodeHeader(buf)
	if err != nil {
		return err
	}
	frame.SetTrailer(trailer)
	return nil
}

func (dec *decoder) decodeHeader(buf bytex.Buffer) (netx.Header, error) {
	var nums uint16
	if err := bytex.ReadUvarint16(buf, &nums); err != nil {
		return nil, err
	}
	if nums == 0 {
		return nil, nil
	}
	header := netx.NewHeader()
	for i := 0; i < int(nums); i++ {
		var key, val string
		if err := bytex.ReadString(buf, &key); err != nil {
			return nil, fmt.Errorf("read header key fail, %+v", key)
		}
		if err := bytex.ReadString(buf, &val); err != nil {
			return nil, fmt.Errorf("read header value fail, %+v", val)
		}
		if len(val) == 0 {
			continue
		}
		values := strings.Split(val, nullStr)
		header.SetValues(key, values)
	}

	return header, nil
}

func getPayload(buf bytex.Buffer) bytex.Buffer {
	buf.Discard()
	if buf.Available() == 0 {
		return nil
	}
	return buf
}
