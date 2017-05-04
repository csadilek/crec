package ingester

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/text/language"

	"mozilla.org/crec/config"
	"mozilla.org/crec/content"
)

func TestMain(m *testing.M) {
	retCode := m.Run()
	tearDown()
	os.Exit(retCode)
}

func tearDown() {
	RemoveAll(config.CreateWithIndexDir("crec-test-index"))
}

func TestGetID(t *testing.T) {
	index := CreateIndex(filepath.FromSlash(os.TempDir()+"/crec-test-index"), "test.bleve")
	id := index.GetID()
	if id == "" {
		t.Error("Index does not have a unique id")
	}
}

func TestGetContent(t *testing.T) {
	index := CreateIndex(filepath.FromSlash(os.TempDir()+"/crec-test-index"), "test.bleve")
	err := index.Add(&content.Content{ID: "0", Summary: "a summary"})
	if err != nil {
		t.Fatal(err)
	}
	content := index.GetContent()
	if len(content) != 1 {
		t.Errorf("Expected content length 1, but got %v", len(content))
	}
}

func TestAddAndQueryContent(t *testing.T) {
	index := CreateIndex(filepath.FromSlash(os.TempDir()+"/crec-test-index"), "test.bleve")
	err := index.Add(&content.Content{ID: "0", Summary: "a summary"})
	if err != nil {
		t.Fatal(err)
	}
	hits, err := index.Query("summary")
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 {
		t.Errorf("Expected exactly one hit, but got %v", len(hits))
	}
}

func TestAddAll(t *testing.T) {
	index := CreateIndex(filepath.FromSlash(os.TempDir()+"/crec-test-index"), "test.bleve")
	err := index.AddAll([]*content.Content{
		&content.Content{ID: "0", Summary: "a summary"},
		&content.Content{ID: "1", Summary: "a summary"}})

	if err != nil {
		t.Fatal(err)
	}
	hits, err := index.Query("summary")
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 2 {
		t.Errorf("Expected exactly two hits, but got %v", len(hits))
	}
}

func TestGetLocalizedContent(t *testing.T) {
	index := CreateIndex(filepath.FromSlash(os.TempDir()+"/crec-test-index"), "test.bleve")
	err := index.Add(&content.Content{ID: "0", Title: "Any", Summary: "a summary"})
	if err != nil {
		t.Fatal(err)
	}
	err = index.Add(&content.Content{ID: "1", Title: "de-at", Summary: "eine kurzfassung", Language: "de", Regions: []string{"AT"}})
	if err != nil {
		t.Fatal(err)
	}

	hits := index.GetLocalizedContent([]language.Tag{language.Make("en-CA")})
	// Should not get de-AT content, but content with ID:0 as it has lang|region|script=any by default
	if len(hits) != 1 {
		t.Fatalf("Expected exactly one hit, but got %v %v", len(hits), hits)
	}
	if hits[0].ID != "0" {
		t.Error("Received invalid content for provide locale")
	}

	hits = index.GetLocalizedContent([]language.Tag{language.Make("de")})
	// Should still not get German content as content is limited to AT region
	if len(hits) != 1 {
		t.Fatalf("Expected exactly one hit, but got %v %v", len(hits), hits)
	}
	if hits[0].ID != "0" {
		t.Error("Received invalid content for provide locale")
	}

	hits = index.GetLocalizedContent([]language.Tag{language.Make("de-AT")})
	// Should get both content flagged as "any" and de-AT
	if len(hits) != 2 {
		t.Errorf("Expected exactly two hits, but got %v %v", len(hits), hits)
	}
}
