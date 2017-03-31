package main

import (
	"log"
	"time"

	"mozilla.org/crec/config"
	"mozilla.org/crec/ingester"
	"mozilla.org/crec/processor"
	"mozilla.org/crec/provider"
	"mozilla.org/crec/server"
)

func main() {
	config := config.Get()
	processors := processor.GetRegistry()

	err := ingester.RemoveAll(config)
	if err != nil {
		log.Println("Failed to delete old content on startup: ", err)
	}

	providers, err := provider.GetProviders(config.GetProviderRegistryDir())
	if err != nil {
		log.Fatal("Failed to read providers from registry: ", err)
	}

	indexer := ingester.Ingest(config, providers, processors)

	s := server.Server{}
	ticker := time.NewTicker(time.Minute * 5)
	go func() {
		for _ = range ticker.C {
			indexer := ingester.Ingest(config, providers, processors)
			s.SetIndexer(indexer)
		}
	}()

	s.Start(config, indexer, providers)
}
