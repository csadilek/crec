package ingester

import (
	"strings"

	"golang.org/x/net/html"

	"github.com/mmcdole/gofeed"
	"mozilla.org/crec/content"
	"mozilla.org/crec/provider"

	"github.com/andybalholm/cascadia"
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
			newc, err := createContentFromFeedItem(provider, item)
			if err != nil {
				return nil, err
			}
			indexer.Add(newc)
			c = append(c, newc)
		}
	}
	indexer.Content = c
	return indexer, nil
}

func createContentFromFeedItem(provider *provider.Provider, item *gofeed.Item) (*content.Content, error) {
	r := strings.NewReader(item.Description)
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	newc := &content.Content{
		Source:  provider.ID,
		Title:   item.Title,
		Link:    item.Link,
		Image:   findImage(doc),
		Summary: findSummary(doc, item.Description),
		Item:    item}
	return newc, nil
}

func findImage(node *html.Node) string {
	img := cascadia.MustCompile("img").MatchFirst(node)
	if img != nil {
		for _, a := range img.Attr {
			if a.Key == "src" {
				if a.Val != "" {
					var src string
					if !strings.HasPrefix(a.Val, "http:") {
						src = "http:" + a.Val
					} else {
						src = a.Val
					}
					return src
				}
			}
		}
	}
	return ""
}

func findSummary(node *html.Node, desc string) string {
	return desc
}
