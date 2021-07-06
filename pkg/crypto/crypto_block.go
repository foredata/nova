package crypto

import (
	"crypto/cipher"
	"fmt"
)

type BlockMode uint8

const (
	CBC BlockMode = iota
	CFB
	CTR
	OFB
	ECB
	GCM
)

type cryptBlocksFunc func(block cipher.Block, iv, src []byte) ([]byte, error)

type blockInfo struct {
	encrypt cryptBlocksFunc
	decrypt cryptBlocksFunc
}

var blockMap = []*blockInfo{
	CBC: {encrypt: encryptCBC, decrypt: decryptCBC},
	CFB: {encrypt: encryptCFB, decrypt: decryptCFB},
	CTR: {encrypt: encryptCTR, decrypt: decryptCTR},
	OFB: {encrypt: encryptOFB, decrypt: decryptOFB},
	ECB: {encrypt: encryptECB, decrypt: decryptECB},
	GCM: {encrypt: encryptGCM, decrypt: decryptGCM},
}

func toSize(mode BlockMode, blockSize int) int {
	if mode == GCM {
		return 12
	} else if mode == ECB {
	}
	switch mode {
	case GCM:
		return 12
	case ECB:
		return 0
	default:
		return blockSize
	}
}

func encryptCBC(block cipher.Block, iv, src []byte) ([]byte, error) {
	dst := make([]byte, len(src))

	bm := cipher.NewCBCEncrypter(block, iv)
	bm.CryptBlocks(dst, src)
	return dst, nil
}

func decryptCBC(block cipher.Block, iv, src []byte) ([]byte, error) {
	dst := make([]byte, len(src))
	bm := cipher.NewCBCDecrypter(block, iv)
	bm.CryptBlocks(dst, src)
	return dst, nil
}

func encryptCFB(block cipher.Block, iv, src []byte) ([]byte, error) {
	dst := make([]byte, len(src))

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(dst, src)
	return dst, nil
}

func decryptCFB(block cipher.Block, iv, src []byte) ([]byte, error) {
	dst := make([]byte, len(src))

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(dst, src)
	return dst, nil
}

func encryptCTR(block cipher.Block, iv, src []byte) ([]byte, error) {
	dst := make([]byte, len(src))

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(dst, src)
	return dst, nil
}

func decryptCTR(block cipher.Block, iv, src []byte) ([]byte, error) {
	dst := make([]byte, len(src))

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(dst, src)
	return dst, nil
}

func encryptOFB(block cipher.Block, iv, src []byte) ([]byte, error) {
	dst := make([]byte, len(src))

	stream := cipher.NewOFB(block, iv)
	stream.XORKeyStream(dst, src)
	return dst, nil
}

func decryptOFB(block cipher.Block, iv, src []byte) ([]byte, error) {
	dst := make([]byte, len(src))

	stream := cipher.NewOFB(block, iv)
	stream.XORKeyStream(dst, src)
	return dst, nil
}

func encryptGCM(block cipher.Block, iv, src []byte) ([]byte, error) {
	if uint64(len(src)) > ((1<<32)-2)*uint64(block.BlockSize()) {
		return nil, fmt.Errorf("crypt AES.Encrypt: plaintext too large for GCM")
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	dst := gcm.Seal(nil, iv, src, nil)
	return dst, nil
}

func decryptGCM(block cipher.Block, iv, src []byte) ([]byte, error) {
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	dst, err := gcm.Open(nil, iv, src, nil)
	if err != nil {
		return nil, err
	}

	return dst, nil
}

func encryptECB(block cipher.Block, iv, src []byte) ([]byte, error) {
	if len(src) == 0 {
		return nil, nil
	}

	blockSize := block.BlockSize()
	if len(src)%blockSize != 0 {
		return nil, fmt.Errorf("crypt/cipher: input not full blocks")
	}

	dst := make([]byte, len(src))
	block.Encrypt(dst, src[:blockSize])
	dst = dst[blockSize:]
	return dst, nil
}

func decryptECB(block cipher.Block, iv, src []byte) ([]byte, error) {
	if len(src) == 0 {
		return nil, nil
	}

	blockSize := block.BlockSize()
	if len(src)%blockSize != 0 {
		return nil, fmt.Errorf("crypt/cipher: input not full blocks")
	}
	dst := make([]byte, len(src))
	block.Decrypt(dst, src)
	dst = dst[blockSize:]
	return dst, nil
}
