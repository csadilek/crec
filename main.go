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
