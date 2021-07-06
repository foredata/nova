package bytex

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// Write .
func Write(w io.Writer, v []byte) error {
	_, err := w.Write(v)
	return err
}

// Writef .
func Writef(w io.Writer, format string, args ...interface{}) error {
	if len(args) > 0 {
		v := fmt.Sprintf(format, args...)
		return Write(w, []byte(v))
	} else {
		return Write(w, []byte(format))
	}
}

// Writev 格式化写入字符串,会先写入长度,Uvarint编码
func Writev(w io.Writer, format string, args ...interface{}) error {
	var v string
	if len(args) > 0 {
		v = fmt.Sprintf(format, args...)
	} else {
		v = format
	}

	return WriteString(w, v)
}

// WriteString 写入带有长度的string
func WriteString(w io.Writer, s string) error {
	size := uint64(len(s))
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, size)
	if err := Write(w, buf[:n]); err != nil {
		return err
	}
	if size > 0 {
		return Write(w, []byte(s))
	}
	return nil
}

// WriteByte 写入字节
func WriteByte(w io.Writer, v byte) error {
	return Write(w, []byte{v})
}

// WriteBool 写入布尔值
func WriteBool(w io.Writer, v bool) error {
	if v {
		return WriteByte(w, 1)
	} else {
		return WriteByte(w, 0)
	}
}

// WriteUvarint64 写入Uvarint编码数据
func WriteUvarint64(w io.Writer, v uint64) error {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, v)
	return Write(w, buf[:n])
}

// WriteVarint64 写入Varint编码数据
func WriteVarint64(w io.Writer, v int64) error {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, v)
	return Write(w, buf[:n])
}

func WriteUvarint32(w io.Writer, v uint32) error {
	buf := make([]byte, binary.MaxVarintLen32)
	n := binary.PutUvarint(buf, uint64(v))
	return Write(w, buf[:n])
}

func WriteVarint32(w io.Writer, v int32) error {
	buf := make([]byte, binary.MaxVarintLen32)
	n := binary.PutVarint(buf, int64(v))
	return Write(w, buf[:n])
}

func WriteUvarint16(w io.Writer, v uint16) error {
	buf := make([]byte, binary.MaxVarintLen16)
	n := binary.PutUvarint(buf, uint64(v))
	return Write(w, buf[:n])
}

func WriteVarint16(w io.Writer, v int16) error {
	buf := make([]byte, binary.MaxVarintLen16)
	n := binary.PutVarint(buf, int64(v))
	return Write(w, buf[:n])
}

//------------------------------------------------
// BigEndian encoding
//------------------------------------------------
func WriteInt16BE(w io.Writer, v int16) error {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(v))
	return Write(w, buf)
}

func WriteInt32BE(w io.Writer, v int32) error {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(v))
	return Write(w, buf)
}

func WriteInt64BE(w io.Writer, v int64) error {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(v))
	return Write(w, buf)
}

func WriteUint16BE(w io.Writer, v uint16) error {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, v)
	return Write(w, buf)
}

func WriteUint32BE(w io.Writer, v uint32) error {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, v)
	return Write(w, buf)
}

func WriteUint64BE(w io.Writer, v uint64) error {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, v)
	return Write(w, buf)
}

func WriteFloat32BE(w io.Writer, v float32) error {
	bits := math.Float32bits(v)
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, bits)
	return Write(w, bytes)
}

func WriteFloat64BE(w io.Writer, v float64) error {
	bits := math.Float64bits(v)
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, bits)
	return Write(w, bytes)
}

//------------------------------------------------
// LittleEndian encoding
//------------------------------------------------
func WriteInt16LE(w io.Writer, v int16) error {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(v))
	return Write(w, buf)
}

func WriteInt32LE(w io.Writer, v int32) error {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(v))
	return Write(w, buf)
}

func WriteInt64LE(w io.Writer, v int64) error {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(v))
	return Write(w, buf)
}

func WriteUint16LE(w io.Writer, v uint16) error {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, v)
	return Write(w, buf)
}

func WriteUint32LE(w io.Writer, v uint32) error {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, v)
	return Write(w, buf)
}

func WriteUint64LE(w io.Writer, v uint64) error {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, v)
	return Write(w, buf)
}

func WriteFloat32LE(w io.Writer, v float32) error {
	bits := math.Float32bits(v)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, bits)
	return Write(w, bytes)
}

func WriteFloat64LE(w io.Writer, v float64) error {
	bits := math.Float64bits(v)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)
	return Write(w, bytes)
}
