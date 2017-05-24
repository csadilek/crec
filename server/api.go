package server

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"strings"

	"log"

	"errors"

	"mozilla.org/crec/config"
	"mozilla.org/crec/provider"
)

// GenerateAPIKey returns an API key for the provider
func GenerateAPIKey(provider string, config *config.Config) string {
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

// GetProviderForAPIKey returns the provider the key was generated for
func GetProviderForAPIKey(key string, config *config.Config) (string, error) {
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

// PrintAPIKeys prints all providers with their corresponding API keys
func PrintAPIKeys(providers provider.Providers, config *config.Config) {
	for provider := range providers {
		apiKey := GenerateAPIKey(provider, config)
		log.Printf("Found provider %v (API key: %v)\n", provider, apiKey)
	}
}
