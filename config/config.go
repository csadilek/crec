package config

import (
	"io/ioutil"
	"path/filepath"

	"log"

	"github.com/BurntSushi/toml"
)

// Config holds all system-wide settings
type Config struct {
	Secret              string
	ServerAddr          string
	ServerContentPath   string
	ServerImportPath    string
	ImportQueueDir      string
	IndexDir            string
	IndexFile           string
	ProviderRegistryDir string
}

// Get returns the configuration based on config.toml, if present.
// Default values are provided for all keys not present, except Secret.
func Get() *Config {
	c := &Config{
		ServerAddr:          ":8080",
		ServerContentPath:   "/crec/content",
		ServerImportPath:    "/crec/import",
		ImportQueueDir:      "import",
		IndexDir:            "index",
		IndexFile:           "crec.bleve",
		ProviderRegistryDir: "crec-registry"}

	bytes, err := ioutil.ReadFile(filepath.FromSlash("config.toml"))
	if err != nil {
		return c
	}

	_, err = toml.Decode(string(bytes), &c)
	if err != nil {
		log.Println("Failed to read provided config (using default settings): ", err)
	}

	return c
}

// GetAddr returns the address and port for starting up the server .e.g :8080
func (c *Config) GetAddr() string {
	return c.ServerAddr
}

// GetContentPath returns the URL path to handle content requests e.g. /crec/content
func (c *Config) GetContentPath() string {
	return c.ServerContentPath
}

// GetImportPath returns the URL path to handle import requests e.g. /crec/import
func (c *Config) GetImportPath() string {
	return c.ServerImportPath
}

// GetImportQueueDir returns the directory path to store imported content e.g. import
func (c *Config) GetImportQueueDir() string {
	return c.ImportQueueDir
}

// GetIndexDir returns the directory path to store indexes e.g. index
func (c *Config) GetIndexDir() string {
	return c.IndexDir
}

// GetIndexFile returns the path to store the index e.g. crec.bleve
func (c *Config) GetIndexFile() string {
	return c.IndexDir
}

// GetProviderRegistryDir returns the directy path containing provider configurations e.g. crec-registry
func (c *Config) GetProviderRegistryDir() string {
	return c.ProviderRegistryDir
}

// GetSecret returns the configured secret to generate API keys
func (c *Config) GetSecret() string {
	return c.Secret
}
