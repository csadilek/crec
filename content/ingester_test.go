package content

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	"net/http"

	"net/http/httptest"

	"mozilla.org/crec/config"
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
	curIndex := CreateIndex(config)

	curIndex.AddItem(&Content{ID: "0", Source: "test"})
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

func TestIngestFromQueue(t *testing.T) {
	config := config.CreateWithIndexDir(filepath.FromSlash(os.TempDir() + "/crec-test-index"))
	providers := provider.Providers{"test": &provider.Provider{ID: "test", Domains: map[string]float32{"d": 0.9}}}

	err := Enqueue(config, []byte(`[{"id":"0"}]`), "test")
	if err != nil {
		t.Errorf("Failed to enqueue content for ingestion: %v", err)
	}

	newIndex := Ingest(config, providers, &Index{})
	content := newIndex.GetContent()

	if len(content) != 1 {
		t.Fatalf("Expected new index to contain content of length 1, but got %v", len(content))
	}
	if content[0].ID != "0" {
		t.Errorf("Invalid content. Expected content with ID 0, but got %v", content[0].ID)
	}
	if content[0].Domains["d"] != 0.9 {
		t.Errorf("Expected provider domains to be present in content")
	}
}

func TestIngestNative(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `[{"id":"0"}]`)
	}))
	defer ts.Close()

	p := &provider.Provider{ID: "test", ContentURL: ts.URL, Domains: map[string]float32{"d": 0.9}}
	config := config.CreateWithIndexDir(filepath.FromSlash(os.TempDir() + "/crec-test-index"))
	index := CreateIndex(config)

	err := ingestNative(p, &http.Client{}, index)
	if err != nil {
		t.Error(err)
	}

	content := index.GetContent()
	if len(content) != 1 {
		t.Fatalf("Expected new index to contain content of length 1, but got %v", len(content))
	}
	if content[0].ID != "0" {
		t.Errorf("Invalid content. Expected content with ID 0, but got %v", content[0].ID)
	}
	if content[0].Domains["d"] != 0.9 {
		t.Errorf("Expected provider domains to be present in content")
	}
}

func TestIngestSyndicationFeed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `<rss><channel><item><guid>0</guid></item></channel></rss>`)
	}))
	defer ts.Close()

	p := &provider.Provider{ID: "test", ContentURL: ts.URL, Domains: map[string]float32{"d": 0.9}}
	config := config.CreateWithIndexDir(filepath.FromSlash(os.TempDir() + "/crec-test-index"))
	index := CreateIndex(config)

	err := ingestSyndicationFeed(p, &http.Client{}, index)
	if err != nil {
		t.Error(err)
	}

	content := index.GetContent()
	if len(content) != 1 {
		t.Fatalf("Expected new index to contain content of length 1, but got %v", len(content))
	}
	if content[0].ID != "0" {
		t.Errorf("Invalid content. Expected content with ID 0, but got %v", content[0].ID)
	}
	if content[0].Domains["d"] != 0.9 {
		t.Errorf("Expected provider domains to be present in content")
	}
}
