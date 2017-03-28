package main

import (
	"log"
	"time"

	"fmt"

	"mozilla.org/crec/ingester"
	"mozilla.org/crec/processor"
	"mozilla.org/crec/provider"
	"mozilla.org/crec/server"
)

func main() {
	providers, err := provider.GetProviders()
	if err != nil {
		log.Fatal("Failed to read content provider registry: ", err)
	}

	indexer := ingester.Ingest(providers, processor.GetRegistry())

	s := server.Server{Addr: ":8080", Path: "/crec/content"}
	ticker := time.NewTicker(time.Minute * 5)
	go func() {
		for _ = range ticker.C {
			indexer := ingester.Ingest(providers, processor.GetRegistry())
			s.SetIndexer(indexer)
		}
	}()
	fmt.Printf("Server listening at %s%s\n", s.Path, s.Addr)
	s.Start(indexer)
}
