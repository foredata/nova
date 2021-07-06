package crypto

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	mrand "math/rand"
	"time"
)

const (
	saltedText       = "salted__"
	saltTextByteSize = len(saltedText)
)

func init() {
	mrand.Seed(time.Now().UnixNano())
}

func randBytes(size int) (r []byte) {
	r = make([]byte, size)
	n, err := rand.Read(r)
	if err != nil || n != size {
		mrand.Read(r)
	}
	return
}

func genSaltHeader(password []byte, keySize int, blockSize int) (header []byte, key, iv []byte) {
	salt := randBytes(saltTextByteSize)
	header = make([]byte, 16)
	copy(header[:], append([]byte(saltedText), salt...))

	key, iv = bytesToKey(salt, password, keySize, keySize+blockSize)
	return
}

func parseSaltHeader(salt []byte, password []byte, keySize int, blockSize int) (key, iv []byte) {
	key, iv = bytesToKey(salt, password, keySize, keySize+blockSize)
	return
}

func getSalt(src []byte) (salt []byte, ok bool) {
	if len(src) >= 16 && bytes.Equal([]byte(saltedText), src[:8]) {
		salt = make([]byte, saltTextByteSize)
		copy(salt[:], src[8:16])
		ok = true
	}

	return
}

func bytesToKey(salt []byte, password []byte, keySize, minimum int) (key, iv []byte) {
	a := append(password, salt...)
	b := md5Sum(a)
	c := append([]byte{}, b...)
	for len(c) < minimum {
		b = md5Sum(append(b, a...))
		c = append(c, b...)
	}
	key = c[:keySize]
	iv = c[keySize:minimum]
	return
}

func md5Sum(data []byte) []byte {
	m := md5.Sum(data)
	return m[:]
}
