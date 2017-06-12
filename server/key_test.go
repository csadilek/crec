package server

import (
	"testing"

	"mozilla.org/crec/config"
)

func TestAPIKeyGen(t *testing.T) {
	config := config.CreateWithSecret("testing-secret-0")
	want := "test-provider"

	apikey := GenerateKey(want, config)
	if got, err := VerifyKey(apikey, config); got != want || err != nil {
		t.Errorf("Failed to verify API key %v\n", err)
	}
}
