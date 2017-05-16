package config

import (
	"strconv"
	"testing"
)

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

func TestGetterMethods(t *testing.T) {
	config := Config{
		serverAddr:                    ":8080",
		serverContentPath:             "/crec/content",
		serverImportPath:              "/crec/import",
		importQueueDir:                "import",
		indexDir:                      "index",
		indexFile:                     "crec.bleve",
		indexRefreshIntervalInMinutes: 5,
		clientCacheMaxAgeInSeconds:    120,
		providerRegistryDir:           "provider-registry",
		templateDir:                   "template",
		secret:                        "dont-do-this"}

	assertEquals(t, config.serverAddr, config.GetAddr())
	assertEquals(t, config.serverContentPath, config.GetContentPath())
	assertEquals(t, config.serverImportPath, config.GetImportPath())
	assertEquals(t, config.importQueueDir, config.GetImportQueueDir())
	assertEquals(t, config.indexDir, config.GetIndexDir())
	assertEquals(t, config.indexFile, config.GetIndexFile())
	assertEquals(t, config.indexRefreshIntervalInMinutes, int64(config.GetIndexRefreshInterval().Minutes()))
	assertEquals(t, strconv.Itoa(int(config.clientCacheMaxAgeInSeconds)), config.GetClientCacheMaxAge())
	assertEquals(t, config.providerRegistryDir, config.GetProviderRegistryDir())
	assertEquals(t, config.templateDir, config.GetTemplateDir())
	assertEquals(t, config.secret, config.GetSecret())
}

func assertEquals(t *testing.T, want interface{}, got interface{}) {
	if want != got {
		t.Errorf("Expected %v, but got %v", want, got)
	}
}
