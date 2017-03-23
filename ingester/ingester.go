package ingester

import (
	"strings"

	"golang.org/x/net/html"

	"github.com/mmcdole/gofeed"
	"mozilla.org/crec/content"
	"mozilla.org/crec/provider"

	"github.com/andybalholm/cascadia"
	"github.com/jaytaylor/html2text"
)

// IngestFrom contacts providers to import content into the system...
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

	var id string
	if item.GUID != "" {
		id = item.GUID
	} else {
		id = item.Link
	}

	newc := &content.Content{
		ID:        id,
		Source:    provider.ID,
		Title:     item.Title,
		Link:      item.Link,
		Image:     findImage(item, doc),
		Summary:   processSummary(doc),
		HTML:      item.Description,
		Tags:      append(item.Categories, provider.Categories...),
		Author:    processAuthor(item),
		Published: item.Published,
		Item:      item}
	return newc, nil
}

func findImage(item *gofeed.Item, node *html.Node) string {

	if item.Image != nil && item.Image.URL != "" {
		return item.Image.URL
	}

	contentExt := item.Extensions["media"]["content"]
	for _, cExt := range contentExt {
		url := cExt.Attrs["url"]
		if url != "" {
			return url
		}
	}

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

func processSummary(node *html.Node) string {
	prepareNode(node)
	text, err := html2text.FromHtmlNode(node)
	if err != nil {
		return ""
	}
	return text
}

func prepareNode(n *html.Node) {
	anchors := findAnchors(n)
	for _, a := range anchors {
		n.RemoveChild(a)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		prepareNode(c)
	}
}

func findAnchors(n *html.Node) []*html.Node {
	anchors := make([]*html.Node, 0)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "a" {
			anchors = append(anchors, c)
		}
	}

	return anchors
}

func processAuthor(item *gofeed.Item) string {
	if item.Author != nil {
		return item.Author.Name
	}
	return ""
}
