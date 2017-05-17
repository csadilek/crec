package ingester

import (
	"io/ioutil"
	"os"
	"path/filepath"
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

	"mozilla.org/crec/config"
	"mozilla.org/crec/processor"
)

// Ingest content from configured providers
func Ingest(config *config.Config, providers provider.Providers, curIndex *Index) *Index {
	CleanUp(config, curIndex)

	index := CreateIndex(config.GetIndexDir(), config.GetIndexFile())
	for _, provider := range providers {
		var err error

		if provider.ContentURL != "" {
			lastUpdated := curIndex.GetProviderLastUpdated(provider.ID)
			nextRefresh := config.GetIndexRefreshInterval()

			if int(time.Now().Add(nextRefresh).Sub(lastUpdated).Minutes()) > provider.MaxContentAge {
				log.Println("Refreshing content from provider " + provider.ID)
				err = ingestFromProvider(provider, index)
				if err == nil {
					index.SetProviderLastUpdated(provider.ID)
				}
			} else {
				log.Println("Reusing content from provider " + provider.ID)
				index.AddAll(curIndex.GetProviderContent(provider.ID))
			}
		} else {
			err = ingestFromQueue(config, provider, index)
		}

		if err != nil {
			index.AddAll(curIndex.GetProviderContent(provider.ID))
			log.Printf("Failed to ingest content from provider %v: %v", provider.ID, err)
		}
	}
	return index
}

// Queue content to be ingested in the next iteration
func Queue(config *config.Config, content []byte, provider string) error {
	path := filepath.Join(config.GetImportQueueDir(), provider)
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	f, err := ioutil.TempFile(path, "import")
	if err != nil {
		return err
	}
	_, err = f.Write(content)
	return err
}

// CleanUp deletes all but the current active index
func CleanUp(config *config.Config, curIndex *Index) {
	indexDirs, err := ioutil.ReadDir(config.GetIndexDir())
	if err != nil {
		log.Println("Failed to clean up old indexes: ", err)
	}

	for _, indexDir := range indexDirs {
		if indexDir.Name() != curIndex.GetID() {
			err = os.RemoveAll(filepath.FromSlash(config.GetIndexDir() + "/" + indexDir.Name()))
			if err != nil {
				log.Println("Failed to clean up old indexes: ", err)
			}
		}
	}
}

func ingestFromProvider(provider *provider.Provider, index *Index) error {
	client := &http.Client{Timeout: time.Duration(time.Second * 5)}
	var err error
	if provider.Native {
		err = ingestNativeJSON(provider, client, index)
	} else {
		err = ingestSyndicationFeed(provider, client, index)
	}
	return err
}

func ingestFromQueue(config *config.Config, provider *provider.Provider, index *Index) error {
	path := filepath.Join(config.GetImportQueueDir(), provider.ID)
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() {
			bytes, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			content := make([]content.Content, 0)
			err = json.Unmarshal(bytes, &content)
			if err != nil {
				return err
			}
			for _, item := range content {
				index.Add(&item)
			}
		}
		return nil
	})
	return err
}

func ingestNativeJSON(provider *provider.Provider, client *http.Client, index *Index) error {
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
		index.Add(&item)
	}

	return nil
}

func ingestSyndicationFeed(provider *provider.Provider, client *http.Client, index *Index) error {
	fp := gofeed.NewParser()
	fp.Client = client
	feed, err := fp.ParseURL(provider.ContentURL)
	if err != nil {
		return err
	}

	for _, item := range feed.Items {
		newc, err := createContentFromFeedItem(provider, item)
		if err != nil {
			return err
		}
		index.Add(newc)
	}

	return nil
}

func createContentFromFeedItem(provider *provider.Provider, item *gofeed.Item) (*content.Content, error) {
	r := strings.NewReader(item.Description)
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	var context = processor.NewHTMLContext(doc)
	for _, processor := range provider.GetProcessors() {
		context, err = processor.Process(context)
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
		Regions:   provider.Regions,
		Language:  provider.Language,
		Script:    provider.Script}
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
