package server

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"strings"

	"mozilla.org/crec/config"

	"log"

	"errors"
)

// GenerateKey returns a consumer (API) key for the given provider
func GenerateKey(provider string, config *config.AppConfig) string {
	plaintext := []byte(provider)

	block, err := aes.NewCipher([]byte(config.GetSecret()))
	if err != nil {
		log.Fatal("Failed to generate API key for provider: "+provider+" ", err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		log.Fatal("Failed to generate API key for provider: " + provider)
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plaintext))
	return strings.Trim(base64.URLEncoding.EncodeToString(ciphertext), "=")
}

// VerifyKey returns the provider the key was generated for, or an error
// if verification failed.
func VerifyKey(key string, config *config.AppConfig) (string, error) {
	if len(key) < aes.BlockSize {
		return "", errors.New("Invalid key length")
	}

	for i := 0; i < len(key)%4; i++ {
		key = key + "="
	}

	apikey, err := base64.URLEncoding.DecodeString(key)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher([]byte(config.GetSecret()))
	if err != nil {
		return "", err
	}
	iv := apikey[:aes.BlockSize]
	provider := apikey[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(provider, provider)
	return string(provider), nil
}
