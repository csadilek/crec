package ingester

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"mozilla.org/crec/config"
	"mozilla.org/crec/content"
	"mozilla.org/crec/provider"
)

func TestIngesterReusesExistingContentOnError(t *testing.T) {
	testIngesterReusesExistingContent(t, &provider.Provider{
		ID:            "test",
		ContentURL:    "invalid-url",
		MaxContentAge: 0})
}

func TestIngesterReusesExistingContentIfNotExpired(t *testing.T) {
	testIngesterReusesExistingContent(t, &provider.Provider{
		ID:            "test",
		ContentURL:    "invalid-url",
		MaxContentAge: math.MaxInt32})
}

func testIngesterReusesExistingContent(t *testing.T, p *provider.Provider) {
	config := config.CreateWithIndexDir(filepath.FromSlash(os.TempDir() + "/crec-test-index"))
	providers := provider.Providers{"test": p}
	curIndex := CreateIndex(config.GetIndexDir(), config.GetIndexFile())

	curIndex.Add(&content.Content{ID: "0", Source: "test"})
	curIndex.SetProviderLastUpdated("test")

	newIndex := Ingest(config, providers, curIndex)
	content := newIndex.GetContent()

	if len(content) != 1 {
		t.Fatalf("Expected new index to contain content of length 1, but got %v", len(content))
	}

	if content[0].ID != "0" {
		t.Errorf("Invalid content. Expected content with ID 0, but got %v", content[0].ID)
	}

	if !newIndex.GetProviderLastUpdated("test").IsZero() {
		t.Error("Content should not have been updated")
	}
}
