package provider

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"log"

	"mozilla.org/crec/config"
)

var regDir string

func TestMain(m *testing.M) {
	before()
	retCode := m.Run()
	tearDown()
	os.Exit(retCode)
}

func before() {
	regDir = filepath.FromSlash(os.TempDir() + "test-provider-registry")
	os.Mkdir(regDir, 0777)

	err := ioutil.WriteFile(filepath.FromSlash(regDir+"/p1.toml"), []byte(createProvider("p1")), 0777)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(filepath.FromSlash(regDir+"/p2.toml"), []byte(createProvider("p2")), 0777)
	if err != nil {
		log.Fatal(err)
	}
}

func tearDown() {
	err := os.RemoveAll(regDir)
	if err != nil {
		log.Print(err)
	}
}

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
		"MaxContentAge=1"
}

func TestGetProviders(t *testing.T) {
	config := config.CreateWithProviderDir(regDir)
	providers, err := GetProviders(config)

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
		MaxContentAge: 1}
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
		MaxContentAge: 1}
	got = providers["p2"]

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Expected provider %v, but got %v", want, got)
	}

}
