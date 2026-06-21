package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

const (
	aes256KeyLen = 32
	gcmNonceSize = 12
)

func GenerateAES256Key() (string, error) {
	key := make([]byte, aes256KeyLen)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", fmt.Errorf("generate aes key failed: %w", err)
	}
	return hex.EncodeToString(key), nil
}

func DeriveAESKey(password string) []byte {
	hash := sha256.Sum256([]byte(password))
	return hash[:]
}

func AESEncryptGCM(plaintext []byte, keyHex string) ([]byte, error) {
	keyBytes, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, fmt.Errorf("decode key failed: %w", err)
	}
	if len(keyBytes) != aes256KeyLen {
		return nil, fmt.Errorf("invalid key length, need %d bytes", aes256KeyLen)
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("new aes cipher failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm failed: %w", err)
	}

	nonce := make([]byte, gcmNonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate nonce failed: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	result := make([]byte, gcmNonceSize+len(ciphertext))
	copy(result[:gcmNonceSize], nonce)
	copy(result[gcmNonceSize:], ciphertext)
	return result, nil
}

func AESDecryptGCM(ciphertext []byte, keyHex string) ([]byte, error) {
	keyBytes, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, fmt.Errorf("decode key failed: %w", err)
	}
	if len(keyBytes) != aes256KeyLen {
		return nil, fmt.Errorf("invalid key length, need %d bytes", aes256KeyLen)
	}

	if len(ciphertext) < gcmNonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce := ciphertext[:gcmNonceSize]
	encryptedData := ciphertext[gcmNonceSize:]

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("new aes cipher failed: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm failed: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, fmt.Errorf("aes decrypt failed: %w", err)
	}
	return plaintext, nil
}

func GenerateExportPassword(length int) string {
	if length <= 0 {
		length = 16
	}
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return GenerateRandomString(length)
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}
