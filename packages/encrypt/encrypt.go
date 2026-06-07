package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"golang.org/x/crypto/argon2"
)

func GenKey() ([]byte, error) {
	x := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, x); err != nil {
		return nil, err
	}
	return x, nil
}
func GenSalt() ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	return salt, nil
}
func DeriveKEK(password string, salt []byte) []byte {
	return argon2.IDKey(
		[]byte(password),
		salt,
		1,       // time cost
		64*1024, // memory (64MB)
		4,       // threads
		32,      // output length (aes 256)
	)
}

func Seal(key []byte, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, data, nil), nil
}
func Open(key []byte, blob []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce, ct := blob[:gcm.NonceSize()], blob[gcm.NonceSize():]

	return gcm.Open(nil, nonce, ct, nil)
}
