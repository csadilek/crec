package provider

import (
	"io/ioutil"
	"os"
	"strings"

	"path/filepath"

	"github.com/BurntSushi/toml"
	"mozilla.org/crec/config"
	"mozilla.org/crec/processor"
)

// Provider represents a content provider.
type Provider struct {
	ID          string                // Unique system-wide identifier of this provider.
	Description string                // Details about this provider.
	URL         string                // URL of the content provider.
	ContentURL  string                // URL to retrieve content, optional in case content is pushed.
	Categories  []string              // Implicit content categories for this provider.
	Processors  []string              // Chain of content processor names, executed in declaration order.
	processors  []processor.Processor // Chain of content processors instances, executed in declaration order.
	Native      bool                  // Native indicates whether or not this provider uses our content format.
}

// Providers is a mapping of provider ID to instance
type Providers map[string]*Provider

// GetProviders returns all registered content providers
func GetProviders(config *config.Config) (Providers, error) {
	return readProvidersFromRegistry(config.GetProviderRegistryDir())
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
	registry := processor.GetRegistry()
	for _, provider := range providers {
		provider.processors = make([]processor.Processor, 0)
		for _, name := range provider.Processors {
			provider.processors = append(provider.processors, registry.GetNewProcessor(name))
		}

		providerMap[provider.ID] = provider
	}

	return providerMap, nil
}

// GetProcessors returns the configured chain of content processors
func (p *Provider) GetProcessors() []processor.Processor {
	return p.processors
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
