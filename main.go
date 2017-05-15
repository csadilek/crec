package main

import (
	"log"
	"time"

	"flag"

	"mozilla.org/crec/config"
	"mozilla.org/crec/ingester"
	"mozilla.org/crec/provider"
	"mozilla.org/crec/server"
)

func main() {
	apiKeys := flag.Bool("apiKeys", false, "Generate and print API keys for providers")
	flag.Parse()

	config := config.Get()
	err := ingester.RemoveAll(config)
	if err != nil {
		log.Println("Failed to delete old content on startup: ", err)
	}
	providers, err := provider.GetProviders(config)
	if err != nil {
		log.Fatal("Failed to read content providers from registry: ", err)
	}
	if *apiKeys {
		printAPIKeys(providers, config)
	}

	index := ingester.Ingest(config, providers, &ingester.Index{})
	server := server.Create(config, index, providers)
	ticker := time.NewTicker(config.GetIndexRefreshInterval())
	go func() {
		for _ = range ticker.C {
			index := ingester.Ingest(config, providers, index)
			server.SetIndex(index)
		}
	}()
	err = server.Start()
	if err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}

func printAPIKeys(providers provider.Providers, config *config.Config) {
	for provider := range providers {
		apiKey := server.GenerateAPIKey(provider, config)
		log.Printf("Found provider %v (API key: %v)\n", provider, apiKey)
	}
}
