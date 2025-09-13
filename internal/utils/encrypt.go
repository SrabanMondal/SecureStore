package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

func Encrypt(plainData, key []byte) ([]byte, []byte, error) {
	if len(key) != 32 {
		return nil, nil, fmt.Errorf("invalid key size: expected 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, plainData, nil)
	return ciphertext, nonce, nil
}

func Decrypt(cipherData, nonce, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("invalid key size: expected 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plainData, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plainData, nil
}
