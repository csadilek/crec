package provider

import (
	"io/ioutil"
	"os"
	"strings"

	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Provider represents a content provider.
type Provider struct {
	ID          string   // Unique system-wide identifier of this provider.
	Description string   // Details about this provider.
	URL         string   // URL of the content provider.
	ContentURL  string   // URL to retrieve content, optional in case content is pushed.
	Categories  []string // Implicit content categories for this provider.
	Processors  []string // Chain of content processors, executed in declaration order.
	Native      bool     // Native indicates whether or not this provider uses our content format.
}

// Providers is a mapping of provider ID to instance
type Providers map[string]*Provider

// GetProviders returns all registered content providers
func GetProviders(providerDir string) (Providers, error) {
	return readProvidersFromRegistry(providerDir)
}

// Using a simple static configuration file based registry for now
func readProvidersFromRegistry(providerDir string) (Providers, error) {
	files, err := ioutil.ReadDir(providerDir)
	if err != nil {
		return nil, err
	}

	files = _filter(files, func(e os.FileInfo) bool {
		return !e.IsDir() && strings.HasSuffix(e.Name(), "toml")
	})

	providers, err := _map(files, func(e os.FileInfo) (*Provider, error) {
		bytes, err := ioutil.ReadFile(filepath.FromSlash(providerDir + "/" + e.Name()))
		if err != nil {
			return nil, err
		}
		var provider Provider
		_, err = toml.Decode(string(bytes), &provider)
		return &provider, err
	})

	providerMap := make(map[string]*Provider)
	for _, provider := range providers {
		providerMap[provider.ID] = provider
	}

	return providerMap, nil
}

func _filter(vs []os.FileInfo, f func(os.FileInfo) bool) []os.FileInfo {
	vsf := make([]os.FileInfo, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func _map(vs []os.FileInfo, f func(os.FileInfo) (*Provider, error)) ([]*Provider, error) {
	vsm := make([]*Provider, len(vs))
	for i, v := range vs {
		e, err := f(v)
		if err != nil {
			return nil, err
		}
		vsm[i] = e
	}
	return vsm, nil
}
