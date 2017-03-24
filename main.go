package main

import (
	"fmt"
	"log"
	"time"

	"mozilla.org/crec/ingester"
	"mozilla.org/crec/provider"
	"mozilla.org/crec/server"
)

func main() {
	providers, err := provider.GetProviders()
	if err != nil {
		log.Fatal("Failed to read content provider registry: ", err)
	}

	print("\nAvailable providers:\n")
	for _, prov := range providers {
		fmt.Printf("%+v\n", *prov)
	}

	indexer := ingester.IngestFrom(providers)

	print("\nAvailable content:\n")
	tags := make(map[string]bool)
	for _, c := range indexer.Content {
		for _, t := range c.Tags {
			tags[t] = true
		}
		fmt.Println(c)
	}

	fmt.Println("\nAvailable tags:")
	for k := range tags {
		println(k)
	}

	s := server.Server{Addr: ":8080", Path: "/crec/content"}
	ticker := time.NewTicker(time.Minute * 5)
	go func() {
		for _ = range ticker.C {
			indexer := ingester.IngestFrom(providers)
			s.SetIndexer(indexer)
		}
	}()
	s.Start(indexer)
}
