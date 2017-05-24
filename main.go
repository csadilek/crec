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

// See: https://docs.google.com/document/d/1PjETbQVZpjtOGkE3sc8XrLUVpMVd02dG24uFqkO3itQ/
func main() {
	apiKeys := flag.Bool("apiKeys", false, "Generate and print API keys for providers")
	flag.Parse()

	config := config.Get()
	providers, err := provider.GetProviders(config)
	if err != nil {
		log.Fatal("Failed to read content providers from registry: ", err)
	}
	if *apiKeys {
		server.PrintAPIKeys(providers, config)
	}

	index := ingester.Ingest(config, providers, &ingester.Index{})
	server := server.Create(config, providers, index)
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
