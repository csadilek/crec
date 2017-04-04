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

	s := server.Server{}
	ticker := time.NewTicker(config.GetIndexRefreshInterval())
	go func() {
		for _ = range ticker.C {
			indexer := ingester.Ingest(config, providers)
			s.SetIndexer(indexer)
		}
	}()

	s.Start(config, indexer, providers)
}
