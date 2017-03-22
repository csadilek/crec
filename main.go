package main

import (
	"fmt"
	"log"

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

	indexer, errC := ingester.IngestFrom(providers)
	if errC != nil {
		log.Fatal("Failed to ingest content from providers: ", err)
	}

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

	fmt.Println("\nStarting server:")
	s := server.Server{Addr: ":8080", Path: "/crec/content", Indexer: indexer}
	s.Start()
}
