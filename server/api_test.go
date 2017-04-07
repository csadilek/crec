package server

import (
	"testing"

	"mozilla.org/crec/config"
)

func TestAPIKeyGen(t *testing.T) {
	config := config.Create("testing-secret-0")
	want := "test-provider"

	apikey := GenerateAPIKey(want, config)
	if got, err := GetProviderForAPIKey(apikey, config); got != want || err != nil {
		t.Errorf("Failed to verify API key %v\n", err)
	}
}
