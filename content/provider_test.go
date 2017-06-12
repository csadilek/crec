package content

import (
	"reflect"
	"testing"
)

var providerDir string

func createProvider(id string) string {
	return "ID=\"" + id + "\"\n" +
		"Description=\"desc\"\n" +
		"URL=\"url\"\n" +
		"ContentURL=\"c_url\"\n" +
		"Processors=[]\n" +
		"Native=true\n" +
		"Language=\"l1\"\n" +
		"Regions=[\"r1\",\"r2\"]\n" +
		"Script=\"s1\"\n" +
		"MaxContentAge=1\n" +
		"Domains={\"bbc.co.uk\"=0.9, \"news.google.com\"=0.8}"
}

func TestGetProviders(t *testing.T) {
	providers, err := GetProviders(&TestConfig{})

	if err != nil {
		t.Fatal(err)
	}

	if len(providers) != 2 {
		t.Errorf("Expected exactly 2 providers, but found %v", len(providers))
	}

	want := &Provider{ID: "p1",
		Description:   "desc",
		URL:           "url",
		ContentURL:    "c_url",
		Native:        true,
		Language:      "l1",
		Processors:    []string{},
		Regions:       []string{"r1", "r2"},
		Script:        "s1",
		MaxContentAge: 1,
		Domains:       map[string]float32{"bbc.co.uk": 0.9, "news.google.com": 0.8}}
	got := providers["p1"]

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Expected provider %v, but got %v", want, got)
	}

	want = &Provider{ID: "p2",
		Description:   "desc",
		URL:           "url",
		ContentURL:    "c_url",
		Native:        true,
		Language:      "l1",
		Processors:    []string{},
		Regions:       []string{"r1", "r2"},
		Script:        "s1",
		MaxContentAge: 1,
		Domains:       map[string]float32{"bbc.co.uk": 0.9, "news.google.com": 0.8}}
	got = providers["p2"]

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Expected provider %v, but got %v", want, got)
	}

}
