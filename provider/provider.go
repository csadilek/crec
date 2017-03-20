package provider

import (
	"io/ioutil"
	"os"
	"strings"

	"path/filepath"

	"github.com/BurntSushi/toml"
)

const registryDir string = "crec-registry"

// Provider represents a content provider.
type Provider struct {
	ID          string // Unique system-wide identifier of this provider.
	Description string // Details about this provider.
	URL         string // URL of the content provider.
	ContentURL  string // URL to retrieve content (TODO need to support pushing content as well)
}

// GetProviders returns all registered content providers
func GetProviders() ([]*Provider, error) {
	return readProvidersFromRegistry()
}

// Using a simple static configuration file based registry for now
func readProvidersFromRegistry() ([]*Provider, error) {
	files, err := ioutil.ReadDir(registryDir)
	if err != nil {
		return nil, err
	}

	files = _filter(files, func(e os.FileInfo) bool {
		return !e.IsDir() && strings.HasSuffix(e.Name(), "toml")
	})

	return _map(files, func(e os.FileInfo) (*Provider, error) {
		bytes, err := ioutil.ReadFile(filepath.FromSlash(registryDir + "/" + e.Name()))
		if err != nil {
			return nil, err
		}
		var provider Provider
		_, err = toml.Decode(string(bytes), &provider)
		return &provider, err
	})
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
