package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

type CryptoService struct {
	secret []byte
}

func NewCryptoService(secret []byte) *CryptoService {
	return &CryptoService{
		secret: secret,
	}
}

func (c *CryptoService) AESEncrypt(data []byte) []byte {
	block, errChiper := aes.NewCipher(c.secret)
	if errChiper != nil {
		panic(errChiper)
	}

	gcm, errGcm := cipher.NewGCM(block)
	if errGcm != nil {
		panic(errGcm)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err)
	}
	chipertext := gcm.Seal(nonce, nonce, data, nil)

	return chipertext
}

func (c *CryptoService) AESDecrypt(data []byte) []byte {
	block, errChiper := aes.NewCipher(c.secret)
	if errChiper != nil {
		panic(errChiper)
	}

	gcm, errGcm := cipher.NewGCM(block)
	if errGcm != nil {
		panic(errGcm)
	}

	nonce := data[:gcm.NonceSize()]
	chipertext := data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, chipertext, nil)
	if err != nil {
		panic(err)
	}
	return plaintext
}
