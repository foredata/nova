package theader

import (
	"encoding/binary"
	"fmt"

	"github.com/foredata/nova/netx"
	"github.com/foredata/nova/pkg/bytex"
)

type decoder struct {
}

func (dec *decoder) Decode(conn netx.Conn, buf bytex.Buffer) (netx.Frame, error) {
	var length, secondword, seqId uint32
	var headerLen uint16

	if err := bytex.ReadUint32BE(buf, &length); err != nil {
		return nil, err
	}
	if length > maxFrameSize {
		return nil, fmt.Errorf("BigFrames not supported: got size %d", length)
	}
	// 数据不足,等待下次解析
	if buf.Available() < int(length+commonHeaderSize) {
		return nil, nil
	}
	if err := bytex.ReadUint32BE(buf, &secondword); err != nil {
		return nil, err
	}

	magic := secondword & magicMask
	// flags := uint16(secondword & flagsMask)
	if magic != headerMagic {
		return nil, fmt.Errorf("invalid magic")
	}
	if err := bytex.ReadUint32BE(buf, &seqId); err != nil {
		return nil, err
	}
	if err := bytex.ReadUint16BE(buf, &headerLen); err != nil {
		return nil, err
	}
	if uint32(headerLen*4) > maxHeaderSize {
		return nil, fmt.Errorf("invalid header length: %d", int64(headerLen*4))
	}
	// Limit the reader for the header so we can't overrun
	headerBuf := buf.ReadN(int(headerLen))

	// read header
	var protoID uint32
	if err := bytex.ReadUvarint32(headerBuf, &protoID); err != nil {
		return nil, err
	}
	if protoID != protoIDBinary && protoID != protoIDCompact {
		return nil, fmt.Errorf("hHeader: invalid protoID, %+v", protoID)
	}

	_, err := dec.readTransforms(headerBuf)
	if err != nil {
		return nil, err
	}
	strHeader, intHeader, err := dec.readInfoHeaders(headerBuf)
	if err != nil {
		return nil, err
	}

	payloadLen := length - commonHeaderSize - uint32(headerLen)*4
	payload := buf.ReadN(int(payloadLen))
	ident := netx.NewIdentifier()
	ident.SeqID = seqId
	strHeader, err = dec.fixIntHeader(intHeader, ident, strHeader)
	if err != nil {
		return nil, err
	}

	frame := netx.NewFrame(netx.FrameTypeHeader, true, 0, ident, strHeader, payload)
	return frame, nil
}

func (dec *decoder) readTransforms(buf bytex.Buffer) ([]TransformID, error) {
	transforms := []TransformID{}
	nums, err := binary.ReadUvarint(buf)
	if err != nil {
		return nil, fmt.Errorf("theader: error reading number of transforms: %s", err.Error())
	}
	for i := 0; i < int(nums); i++ {
		transformID, err := binary.ReadUvarint(buf)
		if err != nil {
			return nil, fmt.Errorf("thheader: error reading transformid: %s", err.Error())
		}
		tid := TransformID(transformID)
		// TODO: check supportedTransforms
		transforms = append(transforms, tid)
	}

	return transforms, nil
}

// readInfoHeaders Read the K/V headers at the end of the header
// This will keep consuming bytes until the buffer returns EOF
func (dec *decoder) readInfoHeaders(buf bytex.Buffer) (netx.Header, IntMap, error) {
	if buf.Available() == 0 {
		return nil, nil, nil
	}
	var strHeader netx.Header
	var intHeader IntMap
	for {
		// this is the last field, read until there is no more padding
		if buf.Available() == 0 {
			break
		}

		infoID, err := binary.ReadUvarint(buf)
		if err != nil {
			return nil, nil, fmt.Errorf("tHeader: error reading infoID: %s", err.Error())
		}
		switch infoIDType(infoID) {
		case infoIDPadding:
			continue
		case infoIDKeyValue:
			h, err := dec.readStringKeyValue(buf)
			if err != nil {
				return nil, nil, err
			}
			strHeader = h
		case infoIDIntKeyValue:
			h, err := dec.readIntKeyValue(buf)
			if err != nil {
				return nil, nil, err
			}
			intHeader = h
		default:
			return nil, nil, fmt.Errorf("tHeader: error reading infoIDType: %#x", infoID)
		}
	}

	return strHeader, intHeader, nil
}

func (dec *decoder) readStringKeyValue(buf bytex.Buffer) (netx.Header, error) {
	nums, err := binary.ReadUvarint(buf)
	if err != nil {
		return nil, fmt.Errorf("tHeader: error reading number of keyvalues: %+v", err)
	}

	if nums == 0 {
		return nil, nil
	}

	headers := netx.NewHeader()
	for i := 0; i < int(nums); i++ {
		var key, val string
		if err := bytex.ReadString(buf, &key); err != nil {
			return nil, fmt.Errorf("tHeader: error reading keyvalue key: %+v", err)
		}
		if err := bytex.ReadString(buf, &val); err != nil {
			return nil, fmt.Errorf("tHeader: error reading keyvalue val: %+v", err)
		}
		headers.Add(key, val)
	}

	return headers, nil
}

func (dec *decoder) readIntKeyValue(buf bytex.Buffer) (IntMap, error) {
	nums, err := binary.ReadUvarint(buf)
	if err != nil {
		return nil, fmt.Errorf("tHeader: error reading number of int keyvalues: %+v", err)
	}
	if nums == 0 {
		return nil, nil
	}
	headers := make(IntMap)
	for i := 0; i < int(nums); i++ {
		var key uint32
		var val string
		if err := bytex.ReadUvarint32(buf, &key); err != nil {
			return nil, fmt.Errorf("tHeader: error reading int kv key: %+v", err)
		}

		if err := bytex.ReadString(buf, &val); err != nil {
			return nil, fmt.Errorf("tHeader: error reading int kv val: %+v", err)
		}

		headers[int(key)] = val
	}

	return headers, nil
}

func (dec *decoder) fixIntHeader(intHeader IntMap, ident *netx.Identifier, strHeader netx.Header) (netx.Header, error) {
	return nil, nil
}
