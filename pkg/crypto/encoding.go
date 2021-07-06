package crypto

import (
	"encoding/base32"
	"encoding/base64"
)

type Encoding interface {
	Encode(src []byte) []byte
	Decode(src []byte) ([]byte, error)
}

func NewBase64Encoding(encoding *base64.Encoding) Encoding {
	return &base64Encoding{encoding: encoding}
}

func NewStdBase64Encoding() Encoding {
	return &base64Encoding{encoding: base64.StdEncoding}
}

type base64Encoding struct {
	encoding *base64.Encoding
}

func (c *base64Encoding) Encode(src []byte) []byte {
	dst := make([]byte, c.encoding.EncodedLen(len(src)))
	c.encoding.Encode(dst, src)
	return dst
}

func (c *base64Encoding) Decode(src []byte) ([]byte, error) {
	dst := make([]byte, c.encoding.DecodedLen(len(src)))
	n, err := c.encoding.Decode(dst, src)
	if err != nil {
		return nil, err
	}
	return dst[:n], nil
}

func NewBase32Encoding(encoding *base32.Encoding) Encoding {
	return &base32Encoding{encoding: encoding}
}

func NewStdBase32Encoding() Encoding {
	return &base32Encoding{encoding: base32.StdEncoding}
}

type base32Encoding struct {
	encoding *base32.Encoding
}

func (c *base32Encoding) Encode(src []byte) []byte {
	dst := make([]byte, c.encoding.EncodedLen(len(src)))
	c.encoding.Encode(dst, src)
	return dst
}

func (c *base32Encoding) Decode(src []byte) ([]byte, error) {
	dst := make([]byte, c.encoding.DecodedLen(len(src)))
	n, err := c.encoding.Decode(dst, src)
	if err != nil {
		return nil, err
	}

	return dst[:n], nil
}
