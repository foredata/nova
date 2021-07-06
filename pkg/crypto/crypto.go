package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
)

const (
	aesSaltKeyByteSize  = 32
	desSaltKeyByteSize  = 8
	des3SaltKeyByteSize = 24
)

// EncryptAES .
func EncryptAES(src, key, iv []byte, mode BlockMode, paddingType PaddingType, encoder Encoding) ([]byte, error) {
	return encrypt(aes.NewCipher, aes.BlockSize, aesSaltKeyByteSize, mode, paddingType, src, key, iv, encoder)
}

func DecryptAES(src, key, iv []byte, mode BlockMode, paddingType PaddingType, decoder Encoding) ([]byte, error) {
	return decrypt(aes.NewCipher, aes.BlockSize, aesSaltKeyByteSize, mode, paddingType, src, key, iv, decoder)
}

func EncryptDES(src, key, iv []byte, mode BlockMode, paddingType PaddingType, encoder Encoding) ([]byte, error) {
	return encrypt(des.NewCipher, des.BlockSize, desSaltKeyByteSize, mode, paddingType, src, key, iv, encoder)
}

func DecryptDES(src, key, iv []byte, mode BlockMode, paddingType PaddingType, decoder Encoding) ([]byte, error) {
	return decrypt(des.NewCipher, des.BlockSize, desSaltKeyByteSize, mode, paddingType, src, key, iv, decoder)
}

func EncryptDES3(src, key, iv []byte, mode BlockMode, paddingType PaddingType, encoder Encoding) ([]byte, error) {
	return encrypt(des.NewTripleDESCipher, des.BlockSize, des3SaltKeyByteSize, mode, paddingType, src, key, iv, encoder)
}

func DecryptDES3(src, key, iv []byte, mode BlockMode, paddingType PaddingType, decoder Encoding) ([]byte, error) {
	return decrypt(des.NewTripleDESCipher, des.BlockSize, des3SaltKeyByteSize, mode, paddingType, src, key, iv, decoder)
}

type newCipherFunc func([]byte) (cipher.Block, error)

func encrypt(newCipher newCipherFunc, blockSize int, keySize int, mode BlockMode, paddingType PaddingType, src, key, iv []byte, encoder Encoding) ([]byte, error) {
	modeInfo := blockMap[mode]
	padding := paddingMap[paddingType].padding

	var header []byte
	if iv == nil && mode != ECB {
		header, key, iv = genSaltHeader(key, keySize, toSize(mode, blockSize))
	}

	block, err := newCipher(key)
	if err != nil {
		return nil, err
	}

	plaintext, err := padding(src, blockSize)
	if err != nil {
		return nil, err
	}

	ciphertext, err := modeInfo.encrypt(block, iv, plaintext)
	if err != nil {
		return nil, err
	}

	if len(header) != 0 {
		ciphertext = append(header, ciphertext...)
	}

	if encoder != nil {
		ciphertext = encoder.Encode(ciphertext)
	}

	return ciphertext, nil
}

func decrypt(newCipher newCipherFunc, blockSize int, keySize int, mode BlockMode, padding PaddingType, src, key, iv []byte, decoder Encoding) ([]byte, error) {
	modeInfo := blockMap[mode]
	unpadding := paddingMap[padding].unpadding

	if decoder != nil {
		var err error
		src, err = decoder.Decode(src)
		if err != nil {
			return nil, err
		}
	}

	if salt, ok := getSalt(src); ok {
		key, iv = parseSaltHeader(salt, key, keySize, toSize(mode, blockSize))
		src = src[16:]
	}

	block, err := newCipher(key)
	if err != nil {
		return nil, err
	}
	dst, err := modeInfo.decrypt(block, iv, src)
	if err != nil {
		return nil, err
	}
	plaintext, err := unpadding(dst, blockSize)
	return plaintext, err
}
