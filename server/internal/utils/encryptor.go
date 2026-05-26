package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
)

type Encryptor struct {
	key []byte
}

func NewEncryptor(secret string) (*Encryptor, error) {
	trimmed := strings.TrimSpace(secret)
	if len(trimmed) < 16 {
		return nil, errors.New("CMDB_SECRET 长度至少 16 位")
	}
	sum := sha256.Sum256([]byte(trimmed))
	return &Encryptor{key: sum[:]}, nil
}

func (e *Encryptor) Encrypt(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	cipherText := gcm.Seal(nil, nonce, []byte(plain), nil)
	payload := append(nonce, cipherText...)
	return base64.StdEncoding.EncodeToString(payload), nil
}

func (e *Encryptor) Decrypt(cipherText string) (string, error) {
	if cipherText == "" {
		return "", nil
	}
	raw, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(raw) <= nonceSize {
		return "", errors.New("invalid ciphertext")
	}
	nonce := raw[:nonceSize]
	encoded := raw[nonceSize:]
	plain, err := gcm.Open(nil, nonce, encoded, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
