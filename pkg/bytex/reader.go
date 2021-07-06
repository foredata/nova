package bytex

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// ReadVarint64 读取varint
func ReadVarint64(r io.ByteReader, out *int64) error {
	v, err := binary.ReadVarint(r)
	if err != nil {
		return err
	}
	*out = v
	return nil
}

// ReadUvarint .
func ReadUvarint64(r io.ByteReader, out *uint64) error {
	v, err := binary.ReadUvarint(r)
	if err != nil {
		return err
	}
	*out = v
	return nil
}

// ReadVarint32 .
func ReadVarint32(r io.ByteReader, out *int32) error {
	v, err := binary.ReadVarint(r)
	if err != nil {
		return err
	}
	*out = int32(v)
	return nil
}

// ReadUvarint32 .
func ReadUvarint32(r io.ByteReader, out *uint32) error {
	v, err := binary.ReadUvarint(r)
	if err != nil {
		return err
	}
	*out = uint32(v)
	return nil
}

// ReadVarint16 .
func ReadVarint16(r io.ByteReader, out *int16) error {
	v, err := binary.ReadVarint(r)
	if err != nil {
		return err
	}
	*out = int16(v)
	return nil
}

// ReadUvarint16 .
func ReadUvarint16(r io.ByteReader, out *uint16) error {
	v, err := binary.ReadUvarint(r)
	if err != nil {
		return err
	}
	*out = uint16(v)
	return nil
}

// ReadString 读取string,要求以长度+content编码
func ReadString(r io.Reader, out *string) error {
	br, ok := r.(io.ByteReader)
	if !ok {
		return fmt.Errorf("[ReadString] cannot convert to io.ByteReader")
	}
	size, err := binary.ReadUvarint(br)
	if err != nil {
		return err
	}
	buf := make([]byte, size)
	n, err := r.Read(buf)
	if uint64(n) != size {
		return err
	}

	*out = string(buf)
	return nil
}

// ReadStringByLen 通过长度读取string
func ReadStringByLen(r io.Reader, len int, out *string) error {
	buf := make([]byte, len)
	n, err := r.Read(buf)
	if n != len {
		return err
	}

	*out = string(buf)
	return nil
}

// ReadByte .
func ReadByte(r io.Reader) (byte, error) {
	if br, ok := r.(io.ByteReader); ok {
		return br.ReadByte()
	}

	buf := make([]byte, 1)
	if _, err := r.Read(buf); err != nil {
		return 0, err
	}

	return buf[0], nil
}

// ReadBool 读取bool
func ReadBool(r io.Reader) (bool, error) {
	b, err := ReadByte(r)
	if err != nil {
		return false, err
	}
	return b == 1, nil
}

//------------------------------------------------
// BigEndian encoding
//------------------------------------------------
// ReadInt16BE .
func ReadInt16BE(r io.Reader, out *int16) error {
	buf := make([]byte, 2)
	if _, err := r.Read(buf); err != nil {
		return err
	}

	*out = int16(binary.BigEndian.Uint16(buf))
	return nil
}

// ReadInt32BE .
func ReadInt32BE(r io.Reader, out *int32) error {
	buf := make([]byte, 4)
	if _, err := r.Read(buf); err != nil {
		return err
	}

	*out = int32(binary.BigEndian.Uint32(buf))
	return nil
}

// ReadInt64BE .
func ReadInt64BE(r io.Reader, out *int64) error {
	buf := make([]byte, 8)
	if _, err := r.Read(buf); err != nil {
		return err
	}

	*out = int64(binary.BigEndian.Uint64(buf))
	return nil
}

// ReadUint16BE .
func ReadUint16BE(r io.Reader, out *uint16) error {
	buf := make([]byte, 2)
	if _, err := r.Read(buf); err != nil {
		return err
	}

	*out = binary.BigEndian.Uint16(buf)
	return nil
}

// ReadUint32BE .
func ReadUint32BE(r io.Reader, out *uint32) error {
	buf := make([]byte, 4)
	if _, err := r.Read(buf); err != nil {
		return err
	}

	*out = binary.BigEndian.Uint32(buf)
	return nil
}

// ReadUint64BE .
func ReadUint64BE(r io.Reader, out *uint64) error {
	buf := make([]byte, 8)
	if _, err := r.Read(buf); err != nil {
		return err
	}

	*out = binary.BigEndian.Uint64(buf)
	return nil
}

// ReadFloat32BE .
func ReadFloat32BE(r io.Reader, out *float32) error {
	buf := make([]byte, 4)
	n, err := r.Read(buf)
	if n < 4 {
		return err
	}

	bits := binary.BigEndian.Uint32(buf)
	*out = math.Float32frombits(bits)
	return nil
}

// ReadFloat64BE 读取float64
func ReadFloat64BE(r io.Reader, out *float64) error {
	buf := make([]byte, 8)
	n, err := r.Read(buf)
	if n < 4 {
		return err
	}

	bits := binary.BigEndian.Uint64(buf)
	*out = math.Float64frombits(bits)
	return nil
}

//------------------------------------------------
// LittleEndian encoding
//------------------------------------------------
// ReadInt16LE .
func ReadInt16LE(r io.Reader, out *int16) error {
	buf := make([]byte, 2)
	if _, err := r.Read(buf); err != nil {
		return err
	}

	*out = int16(binary.LittleEndian.Uint16(buf))
	return nil
}

// ReadInt32LE .
func ReadInt32LE(r io.Reader, out *int32) error {
	buf := make([]byte, 4)
	if _, err := r.Read(buf); err != nil {
		return err
	}

	*out = int32(binary.LittleEndian.Uint32(buf))
	return nil
}

// ReadInt64LE .
func ReadInt64LE(r io.Reader, out *int64) error {
	buf := make([]byte, 8)
	if _, err := r.Read(buf); err != nil {
		return err
	}

	*out = int64(binary.LittleEndian.Uint64(buf))
	return nil
}

// ReadUint16LE .
func ReadUint16LE(r io.Reader, out *uint16) error {
	buf := make([]byte, 2)
	if _, err := r.Read(buf); err != nil {
		return err
	}

	*out = binary.LittleEndian.Uint16(buf)
	return nil
}

// ReadUint32LE .
func ReadUint32LE(r io.Reader, out *uint32) error {
	buf := make([]byte, 4)
	if _, err := r.Read(buf); err != nil {
		return err
	}

	*out = binary.LittleEndian.Uint32(buf)
	return nil
}

// ReadUint64LE .
func ReadUint64LE(r io.Reader, out *uint64) error {
	buf := make([]byte, 8)
	if _, err := r.Read(buf); err != nil {
		return err
	}

	*out = binary.LittleEndian.Uint64(buf)
	return nil
}

// ReadFloat32LE
func ReadFloat32LE(r io.Reader, out *float32) error {
	buf := make([]byte, 4)
	n, err := r.Read(buf)
	if n < 4 {
		return err
	}

	bits := binary.LittleEndian.Uint32(buf)
	*out = math.Float32frombits(bits)
	return nil
}

// ReadFloat64LE .
func ReadFloat64LE(r io.Reader, out *float64) error {
	buf := make([]byte, 8)
	n, err := r.Read(buf)
	if n < 8 {
		return err
	}

	bits := binary.LittleEndian.Uint64(buf)
	*out = math.Float64frombits(bits)
	return nil
}
