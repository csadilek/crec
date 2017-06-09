package content

import (
	"os"
	"path/filepath"
	"testing"

	"mozilla.org/crec/app"
)

func TestGetID(t *testing.T) {
	config := app.CreateConfigWithIndexDir(filepath.FromSlash(os.TempDir() + "/crec-test-index"))
	index := CreateIndex(config)
	id := index.GetID()
	if id == "" {
		t.Error("Index does not have a unique id")
	}
}

func TestGetContent(t *testing.T) {
	config := app.CreateConfigWithIndexDir(filepath.FromSlash(os.TempDir() + "/crec-test-index"))
	index := CreateIndex(config)
	err := index.AddItem(&Content{ID: "0", Summary: "a summary"})
	if err != nil {
		t.Fatal(err)
	}
	content := index.GetContent()
	if len(content) != 1 {
		t.Errorf("Expected content length 1, but got %v", len(content))
	}
}

func TestAddAndQueryContent(t *testing.T) {
	config := app.CreateConfigWithIndexDir(filepath.FromSlash(os.TempDir() + "/crec-test-index"))
	index := CreateIndex(config)
	err := index.AddItem(&Content{ID: "0", Summary: "a summary"})
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

func TestAddAndQueryTitle(t *testing.T) {
	config := app.CreateConfigWithIndexDir(filepath.FromSlash(os.TempDir() + "/crec-test-index"))
	index := CreateIndex(config)
	err := index.AddItem(&Content{ID: "0", Title: "a title", Summary: "a summary"})
	if err != nil {
		t.Fatal(err)
	}
	hits, err := index.Query("title")
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 {
		t.Errorf("Expected exactly one hit, but got %v", len(hits))
	}
}

func TestAddAll(t *testing.T) {
	config := app.CreateConfigWithIndexDir(filepath.FromSlash(os.TempDir() + "/crec-test-index"))
	index := CreateIndex(config)
	err := index.Add([]*Content{
		&Content{ID: "0", Summary: "a summary"},
		&Content{ID: "1", Summary: "a summary"}})

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
	config := app.CreateConfigWithIndexDir(filepath.FromSlash(os.TempDir() + "/crec-test-index"))
	index := CreateIndex(config)
	err := index.AddItem(&Content{ID: "0", Title: "Any", Summary: "a summary"})
	if err != nil {
		t.Fatal(err)
	}
	err = index.AddItem(&Content{
		ID:       "1",
		Title:    "de-at",
		Summary:  "eine kurzfassung",
		Language: "de",
		Regions:  []string{"AT"}})
	if err != nil {
		t.Fatal(err)
	}

	hits := index.GetLocalizedContent("en-CA, en")
	// Should not get de-AT content, but content with ID:0 as it has lang|region|script=any by default
	if len(hits) != 1 {
		t.Fatalf("Expected exactly one hit, but got %v %v", len(hits), hits)
	}
	if hits[0].ID != "0" {
		t.Error("Received invalid content for provided locale")
	}

	hits = index.GetLocalizedContent("de")
	// Should still not get German content as content is limited to AT region
	if len(hits) != 1 {
		t.Fatalf("Expected exactly one hit, but got %v %v", len(hits), hits)
	}
	if hits[0].ID != "0" {
		t.Error("Received invalid content for provided locale")
	}

	hits = index.GetLocalizedContent("de-AT")
	// Should get both content flagged as "any" and de-AT
	if len(hits) != 2 {
		t.Errorf("Expected exactly two hits, but got %v %v", len(hits), hits)
	}
}

func TestGetTaggedContent(t *testing.T) {
	config := app.CreateConfigWithIndexDir(filepath.FromSlash(os.TempDir() + "/crec-test-index"))
	index := CreateIndex(config)
	err := index.Add([]*Content{
		&Content{ID: "0", Tags: []string{"t1"}, Summary: "a summary"},
		&Content{ID: "1", Tags: []string{"t2"}, Summary: "a summary"}})

	if err != nil {
		t.Fatal(err)
	}
	hits := index.GetTaggedContent("t1")

	if len(hits) != 1 {
		t.Errorf("Expected exactly one hit, but got %v", len(hits))
	}
	if hits[0].ID != "0" {
		t.Error("Received invalid content for provided tag")
	}
}
