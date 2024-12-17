package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

// EncryptAES encrypts the given plaintext using AES with CBC mode and a secret key from the environment.
func EncryptAES(plainText string, key string) (string, error) {
	// Convert the key string to a byte slice
	secretKey := []byte(key)
	if len(secretKey) != 16 {
		return "", fmt.Errorf("AES key must be 16, 24, or 32 bytes long")
	}

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}

	// Padding text to be a multiple of block size (16 bytes)
	plainText = pad(plainText, block.BlockSize())

	// Initialization Vector (IV)
	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	// Encrypt
	stream := cipher.NewCBCEncrypter(block, iv)
	stream.CryptBlocks(cipherText[aes.BlockSize:], []byte(plainText))

	// Return hex-encoded ciphertext
	return hex.EncodeToString(cipherText), nil
}

// DecryptAES decrypts the given hex-encoded ciphertext using AES with CBC mode and a secret key from the environment.
func DecryptAES(encryptedText string, key string) (string, error) {
	// Convert the key string to a byte slice
	secretKey := []byte(key)
	if len(secretKey) != 16 && len(secretKey) != 24 && len(secretKey) != 32 {
		return "", fmt.Errorf("AES key must be 16, 24, or 32 bytes long")
	}

	cipherText, err := hex.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}

	// Extract IV and encrypted text
	if len(cipherText) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	// Decrypt
	stream := cipher.NewCBCDecrypter(block, iv)
	stream.CryptBlocks(cipherText, cipherText)

	// Remove padding
	return unpad(string(cipherText)), nil
}

// Pad text to be a multiple of block size
func pad(text string, blockSize int) string {
	padding := blockSize - len(text)%blockSize
	for i := 0; i < padding; i++ {
		text += string(byte(padding))
	}
	return text
}

// Remove padding from decrypted text
func unpad(text string) string {
	padding := text[len(text)-1]
	return text[:len(text)-int(padding)]
}
