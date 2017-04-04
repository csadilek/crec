package config

import (
	"io/ioutil"
	"path/filepath"
	"strconv"

	"log"

	"time"

	"fmt"

	"github.com/BurntSushi/toml"
)

// Config holds all system-wide settings
type Config struct {
	secret                        string
	serverAddr                    string
	serverContentPath             string
	serverImportPath              string
	importQueueDir                string
	indexDir                      string
	indexFile                     string
	indexRefreshIntervalInMinutes int
	providerRegistryDir           string
	clientCacheMaxAgeInSeconds    int
}

// UnmarshalTOML provides a custom "unmarshaller" so we can keep our fields
// private/immutable from outside this package
func (c *Config) UnmarshalTOML(data interface{}) error {
	d := data.(map[string]interface{})
	// EEEK: https://github.com/golang/go/issues/19752
	c.maybeUpdateConfig(d, "Secret", func(val interface{}) { c.secret = val.(string) })
	c.maybeUpdateConfig(d, "ServerAddr", func(val interface{}) { c.serverAddr = val.(string) })
	c.maybeUpdateConfig(d, "ServerContentPath", func(val interface{}) { c.serverContentPath = val.(string) })
	c.maybeUpdateConfig(d, "ServerImportPath", func(val interface{}) { c.serverImportPath = val.(string) })
	c.maybeUpdateConfig(d, "IndexQueueDir", func(val interface{}) { c.importQueueDir = val.(string) })
	c.maybeUpdateConfig(d, "IndexDir", func(val interface{}) { c.indexDir = val.(string) })
	c.maybeUpdateConfig(d, "IndexFile", func(val interface{}) { c.indexFile = val.(string) })
	c.maybeUpdateConfig(d, "IndexRefreshIntervalInMinutes", func(val interface{}) { c.indexRefreshIntervalInMinutes = val.(int) })
	c.maybeUpdateConfig(d, "ProviderRegistryDir", func(val interface{}) { c.providerRegistryDir = val.(string) })
	c.maybeUpdateConfig(d, "ClientCacheMaxAgeInSeconds", func(val interface{}) { c.clientCacheMaxAgeInSeconds = val.(int) })
	return nil
}

func (c *Config) maybeUpdateConfig(data map[string]interface{}, name string, updater func(interface{})) {
	if val, ok := data[name]; ok {
		updater(val)
	}
}

// Get returns the configuration based on config.toml, if present.
// Default values are provided for all keys not present, except Secret.
func Get() *Config {
	c := &Config{
		serverAddr:                    ":8080",
		serverContentPath:             "/crec/content",
		serverImportPath:              "/crec/import",
		importQueueDir:                "import",
		indexDir:                      "index",
		indexFile:                     "crec.bleve",
		indexRefreshIntervalInMinutes: 5,
		clientCacheMaxAgeInSeconds:    120,
		providerRegistryDir:           "crec-registry"}

	bytes, err := ioutil.ReadFile(filepath.FromSlash("config.toml"))
	if err != nil {
		return c
	}

	_, err = toml.Decode(string(bytes), &c)
	if err != nil {
		log.Println("Failed to read provided config (using default settings): ", err)
	}

	fmt.Printf("%v\n", c)
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

// GetIndexRefreshInterval returns the configured refresh interval for the content index
func (c *Config) GetIndexRefreshInterval() time.Duration {
	return time.Minute * time.Duration(int64(c.indexRefreshIntervalInMinutes))
}

// GetClientCacheMaxAge returns the configured cache control max age
func (c *Config) GetClientCacheMaxAge() string {
	return strconv.Itoa(c.clientCacheMaxAgeInSeconds)
}