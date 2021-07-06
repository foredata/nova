package crypto

import (
	"bytes"
	"fmt"
)

func padSize(dataSize, blockSize int) (padding int) {
	padding = blockSize - dataSize%blockSize
	return
}

type PaddingType uint8

const (
	PKCS7 PaddingType = iota
	ANSIX923
	ISO97971
	ISO10126
	ZEROPADDING
	NOPADDING
)

func (p PaddingType) String() string {
	switch p {
	case PKCS7:
		return "PKCS7"
	case ISO97971:
		return "ISO/IEC 9797-1"
	case ANSIX923:
		return "ANSI X.923"
	case ISO10126:
		return "ISO10126"
	case ZEROPADDING:
		return "ZeroPadding"
	case NOPADDING:
		return "NoPadding"
	}
	return ""
}

type paddingFunc func(plaintext []byte, blockSize int) ([]byte, error)

type paddingInfo struct {
	padding   paddingFunc
	unpadding paddingFunc
}

var paddingMap = []paddingInfo{
	PKCS7:       {padding: pkcs7Padding, unpadding: pkcs7UnPadding},
	ANSIX923:    {padding: ansiX923Padding, unpadding: ansiX923UnPadding},
	ISO97971:    {padding: iso97971Padding, unpadding: iso97971UnPadding},
	ISO10126:    {padding: iso10126Padding, unpadding: iso10126UnPadding},
	ZEROPADDING: {padding: zeroPadding, unpadding: zeroUnPadding},
	NOPADDING:   {padding: noPadding, unpadding: noUnPadding},
}

// PKCS7
func pkcs7Padding(plaintext []byte, blockSize int) ([]byte, error) {
	if blockSize < 1 || blockSize > 255 {
		return nil, fmt.Errorf("crypt.PKCS7Padding blockSize is out of bounds: %d", blockSize)
	}
	padding := padSize(len(plaintext), blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plaintext, padtext...), nil
}

func pkcs7UnPadding(ciphertext []byte, blockSize int) ([]byte, error) {
	length := len(ciphertext)
	if length%blockSize != 0 {
		return nil, fmt.Errorf("crypt.PKCS7UnPadding ciphertext's length isn't a multiple of blockSize")
	}
	unpadding := int(ciphertext[length-1])
	if unpadding > blockSize || unpadding <= 0 {
		return nil, fmt.Errorf("crypt.PKCS7UnPadding invalid padding found: %v", unpadding)
	}
	var pad = ciphertext[length-unpadding : length-1]
	for _, v := range pad {
		if int(v) != unpadding {
			return nil, fmt.Errorf("crypt.PKCS7UnPadding invalid padding found")
		}
	}
	return ciphertext[:length-unpadding], nil
}

// ANSI X.923 padding
func ansiX923Padding(plaintext []byte, blockSize int) ([]byte, error) {
	if blockSize < 1 || blockSize > 255 {
		return nil, fmt.Errorf("crypt.AnsiX923Padding blockSize is out of bounds: %d", blockSize)
	}
	padding := padSize(len(plaintext), blockSize)
	padtext := append(bytes.Repeat([]byte{byte(0)}, padding-1), byte(padding))
	return append(plaintext, padtext...), nil
}

func ansiX923UnPadding(ciphertext []byte, blockSize int) ([]byte, error) {
	length := len(ciphertext)
	if length%blockSize != 0 {
		return nil, fmt.Errorf("crypt.AnsiX923UnPadding ciphertext's length isn't a multiple of blockSize")
	}
	unpadding := int(ciphertext[length-1])
	if unpadding > blockSize || unpadding < 1 {
		return nil, fmt.Errorf("crypt.AnsiX923UnPadding invalid padding found: %d", unpadding)
	}
	if length-unpadding < length-2 {
		pad := ciphertext[length-unpadding : length-2]
		for _, v := range pad {
			if int(v) != 0 {
				return nil, fmt.Errorf("crypt.AnsiX923UnPadding invalid padding found")
			}
		}
	}
	return ciphertext[0 : length-unpadding], nil
}

// Zero padding
func zeroPadding(plaintext []byte, blockSize int) ([]byte, error) {
	if blockSize < 1 || blockSize > 255 {
		return nil, fmt.Errorf("crypt.ZeroPadding blockSize is out of bounds: %d", blockSize)
	}
	padding := padSize(len(plaintext), blockSize)
	padtext := bytes.Repeat([]byte{0}, padding)
	return append(plaintext, padtext...), nil
}

func zeroUnPadding(ciphertext []byte, _ int) ([]byte, error) {
	return bytes.TrimRightFunc(ciphertext, func(r rune) bool { return r == rune(0) }), nil
}

// ISO/IEC 9797-1 Padding Method 2
func iso97971Padding(plaintext []byte, blockSize int) ([]byte, error) {
	return zeroPadding(append(plaintext, 0x80), blockSize)
}

func iso97971UnPadding(ciphertext []byte, blockSize int) ([]byte, error) {
	data, err := zeroUnPadding(ciphertext, blockSize)
	if err != nil {
		return nil, err
	}
	return data[:len(data)-1], nil
}

// ISO10126 implements ISO 10126 byte padding. This has been withdrawn in 2007.
func iso10126Padding(plaintext []byte, blockSize int) ([]byte, error) {
	if blockSize < 1 || blockSize > 256 {
		return nil, fmt.Errorf("crypt.ISO10126Padding blockSize is out of bounds: %d", blockSize)
	}
	padding := padSize(len(plaintext), blockSize)
	padtext := append(randBytes(padding-1), byte(padding))
	return append(plaintext, padtext...), nil
}

func iso10126UnPadding(ciphertext []byte, blockSize int) ([]byte, error) {
	length := len(ciphertext)
	if length%blockSize != 0 {
		return nil, fmt.Errorf("crypt.ISO10126UnPadding ciphertext's length isn't a multiple of blockSize")
	}
	unpadding := int(ciphertext[length-1])
	if unpadding > blockSize || unpadding < 1 {
		return nil, fmt.Errorf("crypt.ISO10126UnPadding invalid padding found: %v", unpadding)
	}
	return ciphertext[:length-unpadding], nil
}

func noPadding(plaintext []byte, blockSize int) ([]byte, error) {
	return plaintext, nil
}

func noUnPadding(ciphertext []byte, blockSize int) ([]byte, error) {
	return ciphertext, nil
}
