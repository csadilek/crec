package config

import "testing"

func TestGetConfigReturnsMeaningfulDefaults(t *testing.T) {
	want := Config{
		serverAddr:                    ":8080",
		serverContentPath:             "/crec/content",
		serverImportPath:              "/crec/import",
		importQueueDir:                "import",
		indexDir:                      "index",
		indexFile:                     "crec.bleve",
		indexRefreshIntervalInMinutes: 5,
		clientCacheMaxAgeInSeconds:    120,
		providerRegistryDir:           "provider-registry",
		templateDir:                   "template"}

	got := Get()

	if want != *got {
		t.Errorf("Expected %v, but got %v", want, *got)
	}
}

func TestUnmarshalTOML(t *testing.T) {
	toml := map[string]interface{}{
		"ServerAddr":                    "_serverAddr",
		"ServerContentPath":             "_serverContentPath",
		"ServerImportPath":              "_serverImportPath",
		"ImportQueueDir":                "_importQueueDir",
		"IndexDir":                      "_indexDir",
		"IndexFile":                     "_indexFile",
		"IndexRefreshIntervalInMinutes": int64(1),
		"ClientCacheMaxAgeInSeconds":    int64(2),
		"ProviderRegistryDir":           "_providerRegistryDir",
		"TemplateDir":                   "template"}

	want := Config{
		serverAddr:                    "_serverAddr",
		serverContentPath:             "_serverContentPath",
		serverImportPath:              "_serverImportPath",
		importQueueDir:                "_importQueueDir",
		indexDir:                      "_indexDir",
		indexFile:                     "_indexFile",
		indexRefreshIntervalInMinutes: int64(1),
		clientCacheMaxAgeInSeconds:    int64(2),
		providerRegistryDir:           "_providerRegistryDir",
		templateDir:                   "template"}

	got := &Config{}
	got.UnmarshalTOML(toml)

	if want != *got {
		t.Errorf("Expected %v, but got %v", want, *got)
	}

}
