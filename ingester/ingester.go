package ingester

import (
	"io/ioutil"
	"strings"

	"golang.org/x/net/html"

	"github.com/jaytaylor/html2text"
	"github.com/mmcdole/gofeed"
	"mozilla.org/crec/content"
	"mozilla.org/crec/provider"

	"log"

	"net/http"
	"time"

	"encoding/json"

	"mozilla.org/crec/processor"
)

// Ingest contacts providers to import content into the system...
func Ingest(providers []*provider.Provider, registry *processor.Registry) *Indexer {
	indexer := CreateIndexer()
	client := &http.Client{Timeout: time.Duration(time.Second * 5)}

	for _, provider := range providers {
		var err error
		if provider.Native {
			err = ingestNativeJSON(provider, client, indexer)
		} else {
			err = ingestSyndicationFeed(provider, client, indexer, registry)
		}
		if err != nil {
			log.Println("Failed to ingest content from provider "+provider.ID, err)
			continue
		}
	}
	return indexer
}

func ingestNativeJSON(provider *provider.Provider, client *http.Client, indexer *Indexer) error {
	resp, err := client.Get(provider.ContentURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var content []content.Content
	err = json.Unmarshal(body, &content)
	if err != nil {
		return err
	}

	for _, item := range content {
		indexer.Add(&item)
	}

	return nil
}

func ingestSyndicationFeed(provider *provider.Provider, client *http.Client, indexer *Indexer,
	registry *processor.Registry) error {

	fp := gofeed.NewParser()
	fp.Client = client
	feed, err := fp.ParseURL(provider.ContentURL)
	if err != nil {
		return err
	}

	for _, item := range feed.Items {
		newc, err := createContentFromFeedItem(provider, registry, item)
		if err != nil {
			return err
		}
		indexer.Add(newc)
	}

	return nil
}

func createContentFromFeedItem(provider *provider.Provider, registry *processor.Registry, item *gofeed.Item) (*content.Content, error) {
	r := strings.NewReader(item.Description)
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	var context = processor.NewHTMLContext(doc)
	for _, name := range provider.Processors {
		context, err = registry.GetNewProcessor(name).Process(context)
		if err != nil {
			return nil, err
		}
	}

	summary, err := html2text.FromHtmlNode(context.Content.(*html.Node))
	if err != nil {
		return nil, err
	}

	newc := &content.Content{
		ID:        findID(item),
		Source:    provider.ID,
		Title:     item.Title,
		Link:      item.Link,
		Image:     findImage(item, context),
		Summary:   summary,
		HTML:      item.Description,
		Tags:      append(item.Categories, provider.Categories...),
		Author:    processAuthor(item),
		Published: item.Published,
		Item:      item}
	return newc, nil
}

func findImage(item *gofeed.Item, context *processor.Context) string {
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

	return context.Result["image"]
}

func processAuthor(item *gofeed.Item) string {
	if item.Author != nil {
		return item.Author.Name
	}
	return ""
}

func findID(item *gofeed.Item) string {
	if item.GUID != "" {
		return item.GUID
	}
	return item.Link
}
