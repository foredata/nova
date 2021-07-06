package rpc

import "encoding/binary"

const (
	magicWord = 0x4b47

	magicMask    = 0x8000 // 标记是否有magic number
	versionMask  = 0x4000 // 标记是否有version字段
	frameEndMask = 0x2000 // 标记是否最后一帧
	cmdIdMask    = 0x1000 // 标记是否使用cmdId,否则使用URI

	// 偏移
	frameTypeShift = 10
	msgTypeShift   = 8
	//
	maxLengthBytes = binary.MaxVarintLen32
	maxHeaderNum   = 65535
)

var nullStr = string([]byte{0})

func hasFlag(f, mask uint16) bool {
	return (f & mask) != 0
}

func setFlag(f *uint16, mask uint16) {
	*f |= mask
}

func get2Bits(f, shift uint16) uint16 {
	return (f >> shift) & 0x03
}

func set2Bits(f *uint16, value, shift uint16) {
	*f |= (value & 0x03) << shift
}

// func get4Bits(f, shift uint16) uint16 {
// 	return (f >> shift) & 0x0F
// }

// func set4Bits(f *uint16, value, shift uint16) {
// 	*f |= (value & 0x0F) << shift
// }
