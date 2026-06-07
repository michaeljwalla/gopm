package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"errors"
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

func SealField(dek []byte, field sql.NullString) ([]byte, error) {
	var plaintext []byte
	if field.Valid {
		plaintext = append([]byte{0x01}, []byte(field.String)...)
	} else {
		var sizeBuf [1]byte
		if _, err := io.ReadFull(rand.Reader, sizeBuf[:]); err != nil {
			return nil, err
		}
		paddingLen := 31 + int(sizeBuf[0])%33
		padding := make([]byte, paddingLen)
		if _, err := io.ReadFull(rand.Reader, padding); err != nil {
			return nil, err
		}
		plaintext = append([]byte{0x00}, padding...)
	}
	return Seal(dek, plaintext)
}
func OpenField(dek []byte, blob []byte) (sql.NullString, error) {
	plaintext, err := Open(dek, blob)
	if err != nil {
		return sql.NullString{}, err
	}
	if len(plaintext) == 0 || plaintext[0] > 0x01 {
		return sql.NullString{}, errors.New("malformed or missing field")
	}
	if plaintext[0] == 0x01 {
		return sql.NullString{String: string(plaintext[1:]), Valid: true}, nil
	}
	return sql.NullString{}, nil
}
