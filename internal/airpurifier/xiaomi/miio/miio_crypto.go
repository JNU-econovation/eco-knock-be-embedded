package miio

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
)

func Encrypt(plaintext []byte, token []byte) ([]byte, error) {
	key, iv := deriveKeyIV(token)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	padded := pkcs7Pad(plaintext, aes.BlockSize)
	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, padded)

	return ciphertext, nil
}

func Decrypt(ciphertext []byte, token []byte) ([]byte, error) {
	key, iv := deriveKeyIV(token)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, ErrInvalidResponse
	}

	plaintext := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plaintext, ciphertext)

	unpadded, err := pkcs7Unpad(plaintext, aes.BlockSize)
	if err != nil {
		return nil, err
	}

	return bytes.TrimRight(unpadded, "\x00"), nil
}

func deriveKeyIV(token []byte) ([]byte, []byte) {
	keyChecksum := md5.Sum(token) //nolint:gosec
	key := append([]byte(nil), keyChecksum[:]...)

	ivInput := append(append([]byte{}, key...), token...)
	ivChecksum := md5.Sum(ivInput) //nolint:gosec
	iv := append([]byte(nil), ivChecksum[:]...)

	return key, iv
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	if padding == 0 {
		padding = blockSize
	}

	return append(data, bytes.Repeat([]byte{byte(padding)}, padding)...)
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, ErrInvalidResponse
	}

	padding := int(data[len(data)-1])
	if padding == 0 || padding > blockSize || padding > len(data) {
		return nil, ErrInvalidResponse
	}

	for _, value := range data[len(data)-padding:] {
		if int(value) != padding {
			return nil, ErrInvalidResponse
		}
	}

	return data[:len(data)-padding], nil
}
