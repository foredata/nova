package crypto_test

import (
	"testing"

	"github.com/foredata/nova/pkg/crypto"
)

func TestAES(t *testing.T) {
	key := []byte(`15234c27ef5da06b`)
	text := []byte("hello")
	encoding := crypto.NewStdBase64Encoding()
	ciphertext, err := crypto.EncryptAES(text, key, nil, crypto.CBC, crypto.PKCS7, encoding)
	t.Log(string(ciphertext), err)
	plaintext, err := crypto.DecryptAES(ciphertext, key, nil, crypto.CBC, crypto.PKCS7, encoding)
	t.Log(string(plaintext), err)
}

func TestDES(t *testing.T) {

}

func TestDES3(t *testing.T) {

}
