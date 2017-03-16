package ingester

import (
	"github.com/mmcdole/gofeed"
	"mozilla.org/cas/content"
	"mozilla.org/cas/provider"
)

// IngestFrom contacts the providers and imports content into the system...
func IngestFrom(providers []*provider.Provider) (*Indexer, error) {
	c := make([]*content.Content, 0)
	indexer := CreateIndexer()
	fp := gofeed.NewParser()
	for _, provider := range providers {
		feed, err := fp.ParseURL(provider.ContentURL)
		if err != nil {
			return nil, err
		}
		for _, item := range feed.Items {
			newc := &content.Content{Source: provider.ID, Item: item}
			indexer.Add(newc)
			c = append(c, newc)
		}
	}
	indexer.Content = c
	return indexer, nil
}
