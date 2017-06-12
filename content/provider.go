package content

import (
	"io/ioutil"
	"os"
	"strings"

	"path/filepath"

	"github.com/BurntSushi/toml"
	"mozilla.org/crec/content/processor"
)

// Provider represents a content provider.
type Provider struct {
	// Unique system-wide identifier of this provider.
	ID string

	// Details about this provider.
	Description string

	// URL of this content provider.
	URL string

	// URL to retrieve content, optional in case content is pushed.
	ContentURL string

	// Implicit content categories for this provider.
	Categories []string

	// Chain of content processor names, executed in declaration order.
	Processors []string

	// Chain of content processors instances, executed in declaration order.
	processors []processor.Processor

	// Native indicates whether or not this provider uses our content format.
	Native bool

	// Specifies the default applicable regions for this provider’s content.
	// If omitted, content will be considered for all regions, unless
	// specified otherwise in content.
	Regions []string

	// Specifies the default applicable language for this provider’s content.
	// If omitted, content will be considered for all languages, unless
	// specified otherwise in content.
	Language string

	// Specifies the default applicable script for this provider’s content.
	// If omitted, content will be considered for all scripts, unless
	// specified otherwise in content.
	Script string

	// Specifies the time in minutes after which this provider's content
	// should be refreshed.
	MaxContentAge int

	// Specifies the default domain similarities of this provider. The domain
	// name is used as key, the weight as value. This can be used on the client
	// to map content of this provider to specific user interests i.e. based on
	// their browsing history.
	Domains map[string]float32
}

// Providers is a mapping of provider IDs to instances
type Providers map[string]*Provider

// GetProviders returns all registered content providers
func GetProviders(config Config) (Providers, error) {
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

	if err != nil {
		return nil, err
	}

	providerMap := make(map[string]*Provider)
	registry := processor.GetRegistry()
	for _, provider := range providers {
		if len(provider.Processors) > 0 {
			provider.processors = make([]processor.Processor, 0)
			for _, name := range provider.Processors {
				provider.processors = append(provider.processors, registry.GetNewProcessor(name))
			}
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
