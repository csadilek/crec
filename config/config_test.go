package config

import (
	"strconv"
	"testing"
)

func TestGetConfigReturnsMeaningfulDefaults(t *testing.T) {
	want := AppConfig{
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
		"FullTextIndex":                 true,
		"FullTextIndexDir":              "_indexDir",
		"FullTextIndexFile":             "_indexFile",
		"IndexRefreshIntervalInMinutes": int64(1),
		"ClientCacheMaxAgeInSeconds":    int64(2),
		"ProviderRegistryDir":           "_providerRegistryDir",
		"TemplateDir":                   "template",
		"Locales":                       "en, en-US"}

	want := AppConfig{
		serverAddr:                    "_serverAddr",
		serverContentPath:             "_serverContentPath",
		serverImportPath:              "_serverImportPath",
		importQueueDir:                "_importQueueDir",
		fullTextIndex:                 true,
		fullTextIndexDir:              "_indexDir",
		fullTextIndexFile:             "_indexFile",
		indexRefreshIntervalInMinutes: int64(1),
		clientCacheMaxAgeInSeconds:    int64(2),
		providerRegistryDir:           "_providerRegistryDir",
		templateDir:                   "template",
		locales:                       "en, en-US"}

	got := &AppConfig{}
	got.UnmarshalTOML(toml)

	if want != *got {
		t.Errorf("Expected %v, but got %v", want, *got)
	}

}

func TestGetterMethods(t *testing.T) {
	config := AppConfig{
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
		secret:                        "dont-do-this",
		locales:                       "en, en-US"}

	assertEquals(t, config.serverAddr, config.GetAddr())
	assertEquals(t, config.serverContentPath, config.GetContentPath())
	assertEquals(t, config.serverImportPath, config.GetImportPath())
	assertEquals(t, config.importQueueDir, config.GetImportQueueDir())
	assertEquals(t, config.fullTextIndex, config.FullTextIndexActive())
	assertEquals(t, config.fullTextIndexDir, config.GetFullTextIndexDir())
	assertEquals(t, config.fullTextIndexFile, config.GetFullTextIndexFile())
	assertEquals(t, config.indexRefreshIntervalInMinutes, int64(config.GetIndexRefreshInterval().Minutes()))
	assertEquals(t, strconv.Itoa(int(config.clientCacheMaxAgeInSeconds)), config.GetClientCacheMaxAge())
	assertEquals(t, config.providerRegistryDir, config.GetProviderRegistryDir())
	assertEquals(t, config.templateDir, config.GetTemplateDir())
	assertEquals(t, config.secret, config.GetSecret())
	assertEquals(t, config.locales, config.GetLocales())
}

func TestCreateMethods(t *testing.T) {
	c := CreateWithSecret("secret")
	if c.secret != "secret" {
		t.Error("Failed to create config with secret")
	}

	c = Create("secret", "templateDir", "importQueueDir", "indexDir", "indexFile")
	if c.secret != "secret" {
		t.Error("Failed to create config with secret")
	}
	if c.templateDir != "templateDir" {
		t.Error("Failed to create config with template dir")
	}
	if c.importQueueDir != "importQueueDir" {
		t.Error("Failed to create config with import queue dir")
	}
	if c.fullTextIndexDir != "indexDir" {
		t.Error("Failed to create config with index dir")
	}
	if c.fullTextIndexFile != "indexFile" {
		t.Error("Failed to create config with index file")
	}
}

func assertEquals(t *testing.T, want interface{}, got interface{}) {
	if want != got {
		t.Errorf("Expected %v, but got %v", want, got)
	}
}
