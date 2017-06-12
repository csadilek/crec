package content

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	before()
	retCode := m.Run()
	tearDown()
	os.Exit(retCode)
}

type TestConfig struct{}

func (t *TestConfig) GetFullTextIndexDir() string {
	return filepath.FromSlash(os.TempDir() + "/crec-test-index")
}
func (t *TestConfig) GetFullTextIndexFile() string {
	return "test_crec.bleve"
}
func (t *TestConfig) GetImportQueueDir() string {
	return "import"
}
func (t *TestConfig) GetIndexRefreshInterval() time.Duration {
	return time.Minute * time.Duration(int64(5))
}
func (t *TestConfig) FullTextIndexActive() bool {
	return true
}
func (t *TestConfig) GetLocales() string {
	return ""
}
func (t *TestConfig) GetProviderRegistryDir() string {
	return providerDir
}

func before() {
	providerDir = filepath.FromSlash(os.TempDir() + "test-provider-registry")
	os.Mkdir(providerDir, 0777)

	err := ioutil.WriteFile(filepath.FromSlash(providerDir+"/p1.toml"), []byte(createProvider("p1")), 0777)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(filepath.FromSlash(providerDir+"/p2.toml"), []byte(createProvider("p2")), 0777)
	if err != nil {
		log.Fatal(err)
	}
}

func tearDown() {
	config := &TestConfig{}
	cleanUp(config, &Index{})
	os.RemoveAll(config.GetImportQueueDir())

	if providerDir != "" {
		os.RemoveAll(providerDir)
	}
}

func TestFilterContent(t *testing.T) {
	content := []*Content{{ID: "0"}, {ID: "1"}, {ID: "2"}}

	content = Filter(content, func(c *Content) bool {
		return c.ID == "1"
	})
	if len(content) != 1 {
		t.Error("Expected exactly one item")
	}
	if content[0].ID != "1" {
		t.Error("Filtered out incorrect item")
	}
}

func TestAnyTagFilter(t *testing.T) {
	content := []*Content{{Tags: []string{"t1", "t2", "t3"}}}

	filtered := Filter(content, AnyTagFilter(map[string]bool{"t4": true}))
	if len(filtered) > 0 {
		t.Error("Should not have found match")
	}

	filtered = Filter(content, AnyTagFilter(map[string]bool{"t2": true}))
	if len(filtered) != 1 {
		t.Errorf("Should have found exactly one match, but found %v", len(filtered))
	}
}

func TestAllTagFilter(t *testing.T) {
	content := []*Content{{Tags: []string{"t1", "t2", "t3"}}}

	filtered := Filter(content, AllTagFilter([]string{"t2", "t4"}))
	if len(filtered) > 0 {
		t.Error("Should not have found match")
	}

	filtered = Filter(content, AllTagFilter([]string{"t1", "t2", "t3"}))
	if len(filtered) != 1 {
		t.Errorf("Should have found exactly one match, but found %v", len(filtered))
	}
}

func TestTransformContent(t *testing.T) {
	content := []*Content{{ID: "0"}, {ID: "1"}, {ID: "2"}}

	content = Transform(content, func(c Content) *Content {
		c.Title = "Transformed"
		return &c
	})
	if len(content) != 3 {
		t.Error("Length of array should not have changed")
	}
	for i := range content {
		if content[i].Title != "Transformed" {
			t.Errorf("Failed to transform item at index: %v", i)
		}
	}
}

func TestTransformContentIsThreadSafe(t *testing.T) {
	content := []*Content{{ID: "0"}, {ID: "1"}, {ID: "2"}}

	Transform(content, func(c Content) *Content {
		c.ID = "Transformed"
		return &c
	})
	if len(content) != 3 {
		t.Error("Length of array should not have changed")
	}
	for i := range content {
		if content[i].ID == "Transformed" {
			t.Errorf("Tranform should NOT change original array")
		}
	}
}
