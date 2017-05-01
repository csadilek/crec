package main

import (
	"log"
	"time"

	"mozilla.org/crec/config"
	"mozilla.org/crec/ingester"
	"mozilla.org/crec/provider"
	"mozilla.org/crec/server"
)

func main() {
	config := config.Get()

	err := ingester.RemoveAll(config)
	if err != nil {
		log.Println("Failed to delete old content on startup: ", err)
	}

	providers, err := provider.GetProviders(config)
	if err != nil {
		log.Fatal("Failed to read content providers from registry: ", err)
	}

	indexer := ingester.Ingest(config, providers)

	server := server.Create(config, indexer, providers)

	ticker := time.NewTicker(config.GetIndexRefreshInterval())
	go func() {
		for _ = range ticker.C {
			indexer := ingester.Ingest(config, providers)
			server.SetIndexer(indexer)
		}
	}()

	err = server.Start()
	if err != nil {
		log.Fatal("Server failed to start: ", err)
	}
}
