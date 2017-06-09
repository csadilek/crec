package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"log"

	"time"

	"github.com/BurntSushi/toml"
)

// Config holds all system-wide settings
type Config struct {
	secret                        string
	serverAddr                    string
	serverContentPath             string
	serverImportPath              string
	importQueueDir                string
	fullTextIndex                 bool
	fullTextIndexDir              string
	fullTextIndexFile             string
	indexRefreshIntervalInMinutes int64
	providerRegistryDir           string
	clientCacheMaxAgeInSeconds    int64
	templateDir                   string
	locales                       string
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
	c.maybeUpdateConfig(d, "ImportQueueDir", func(val interface{}) { c.importQueueDir = val.(string) })
	c.maybeUpdateConfig(d, "FullTextIndex", func(val interface{}) { c.fullTextIndex = val.(bool) })
	c.maybeUpdateConfig(d, "FullTextIndexDir", func(val interface{}) { c.fullTextIndexDir = val.(string) })
	c.maybeUpdateConfig(d, "FullTextIndexFile", func(val interface{}) { c.fullTextIndexFile = val.(string) })
	c.maybeUpdateConfig(d, "IndexRefreshIntervalInMinutes", func(val interface{}) { c.indexRefreshIntervalInMinutes = val.(int64) })
	c.maybeUpdateConfig(d, "ProviderRegistryDir", func(val interface{}) { c.providerRegistryDir = val.(string) })
	c.maybeUpdateConfig(d, "ClientCacheMaxAgeInSeconds", func(val interface{}) { c.clientCacheMaxAgeInSeconds = val.(int64) })
	c.maybeUpdateConfig(d, "TemplateDir", func(val interface{}) { c.templateDir = val.(string) })
	c.maybeUpdateConfig(d, "Locales", func(val interface{}) { c.locales = val.(string) })
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
		fullTextIndex:                 true,
		fullTextIndexDir:              "index",
		fullTextIndexFile:             "crec.bleve",
		indexRefreshIntervalInMinutes: 5,
		clientCacheMaxAgeInSeconds:    120,
		providerRegistryDir:           "provider-registry",
		templateDir:                   "template",
		locales:                       "en, en-US"}

	port := os.Getenv("PORT")
	if port != "" {
		c.serverAddr = ":" + port
	}

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

// FullTextIndexActive returns true if a full-text index should be created, otherwise false.
func (c *Config) FullTextIndexActive() bool {
	return c.fullTextIndex
}

// GetFullTextIndexDir returns the directory path to store indexes e.g. index
func (c *Config) GetFullTextIndexDir() string {
	return c.fullTextIndexDir
}

// GetFullTextIndexFile returns the path to store the index e.g. crec.bleve
func (c *Config) GetFullTextIndexFile() string {
	return c.fullTextIndexFile
}

// GetProviderRegistryDir returns the directy path containing provider configurations e.g. provider-registry
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
	return strconv.Itoa(int(c.clientCacheMaxAgeInSeconds))
}

// GetTemplateDir returns the configured html template directory
func (c *Config) GetTemplateDir() string {
	return c.templateDir
}

// GetLocales returns the configured default locales of this node
func (c *Config) GetLocales() string {
	return c.locales
}

// Create returns a config instance with the provided parameters
func Create(secret string, templateDir string, importQueueDir string,
	fullTextIndexDir string, fullTextIndexFile string) *Config {

	config := Get()
	config.secret = secret
	config.templateDir = templateDir
	config.importQueueDir = importQueueDir
	config.fullTextIndexDir = fullTextIndexDir
	config.fullTextIndexFile = fullTextIndexFile
	return config
}

// CreateWithSecret returns a config instance with the provided secret
func CreateWithSecret(secret string) *Config {
	config := Get()
	config.secret = secret
	return config
}

// CreateWithIndexDir returns a config instance with the provided index dir
func CreateWithIndexDir(indexDir string) *Config {
	config := Get()
	config.fullTextIndexDir = indexDir
	return config
}

// CreateWithProviderDir returns a config instance with the given provider dir
func CreateWithProviderDir(providerDir string) *Config {
	config := Get()
	config.providerRegistryDir = providerDir
	return config
}
