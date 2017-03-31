package config

import (
	"io/ioutil"
	"path/filepath"

	"log"

	"github.com/BurntSushi/toml"
)

// Config holds all system-wide settings
type Config struct {
	secret              string
	serverAddr          string
	serverContentPath   string
	serverImportPath    string
	importQueueDir      string
	indexDir            string
	indexFile           string
	providerRegistryDir string
}

// Get returns the configuration based on config.toml, if present.
// Default values are provided for all keys not present, except Secret.
func Get() *Config {
	c := &Config{
		serverAddr:          ":8080",
		serverContentPath:   "/crec/content",
		serverImportPath:    "/crec/import",
		importQueueDir:      "import",
		indexDir:            "index",
		indexFile:           "crec.bleve",
		providerRegistryDir: "crec-registry"}

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
	return c.serverAddr
}

// GetContentPath returns the URL path to handle content requests e.g. /crec/content
func (c *Config) GetContentPath() string {
	return c.serverContentPath
}

// GetImportPath returns the URL path to handle import requests e.g. /crec/import
func (c *Config) GetImportPath() string {
	return c.serverImportPath
}

// GetImportQueueDir returns the directory path to store imported content e.g. import
func (c *Config) GetImportQueueDir() string {
	return c.importQueueDir
}

// GetIndexDir returns the directory path to store indexes e.g. index
func (c *Config) GetIndexDir() string {
	return c.indexDir
}

// GetIndexFile returns the path to store the index e.g. crec.bleve
func (c *Config) GetIndexFile() string {
	return c.indexDir
}

// GetProviderRegistryDir returns the directy path containing provider configurations e.g. crec-registry
func (c *Config) GetProviderRegistryDir() string {
	return c.providerRegistryDir
}

// GetSecret returns the configured secret to generate API keys
func (c *Config) GetSecret() string {
	return c.secret
}
