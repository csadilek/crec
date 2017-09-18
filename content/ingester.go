package content

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/net/html"

	"github.com/jaytaylor/html2text"
	"github.com/mmcdole/gofeed"

	"log"

	"net/http"
	"time"

	"encoding/json"

	"mozilla.org/crec/content/processor"
)

// Ingest content from configured providers
func Ingest(config Config, providers Providers, curIndex *Index) *Index {
	cleanUp(config, curIndex)

	index := CreateIndex(config)

	var wg sync.WaitGroup
	wg.Add(len(providers))
	for _, p := range providers {
		go func(provider *Provider) {
			defer wg.Done()
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
					index.Add(curIndex.GetProviderContent(provider.ID))
				}
			} else {
				err = ingestFromQueue(config, provider, index)
			}

			if err != nil {
				index.Add(curIndex.GetProviderContent(provider.ID))
				log.Printf("Failed to refresh content from provider %v: %v", provider.ID, err)
			}
		}(p)
	}
	wg.Wait()

	index.PreLoadLocales(config.GetLocales())
	log.Println("Indexing complete")
	return index
}

// Enqueue writes content to the disc to be ingested in the next indexing iteration
func Enqueue(config Config, content []byte, provider string) error {
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

// cleanUp deletes all but the current active index
func cleanUp(config Config, curIndex *Index) {
	indexDirs, _ := ioutil.ReadDir(config.GetFullTextIndexDir())
	for _, indexDir := range indexDirs {
		if indexDir.Name() != curIndex.GetID() {
			err := os.RemoveAll(filepath.FromSlash(config.GetFullTextIndexDir() + "/" + indexDir.Name()))
			if err != nil {
				log.Println("Failed to clean up old indexes: ", err)
			}
		}
	}
}

func ingestFromProvider(provider *Provider, index *Index) error {
	client := &http.Client{Timeout: time.Duration(time.Second * 5)}
	var err error
	if provider.Native {
		err = ingestNative(provider, client, index)
	} else {
		err = ingestSyndicationFeed(provider, client, index)
	}
	return err
}

func ingestFromQueue(config Config, provider *Provider, index *Index) error {
	path := filepath.Join(config.GetImportQueueDir(), provider.ID)
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() {
			bytes, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			err = ingestJSON(bytes, provider, index)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func ingestNative(provider *Provider, client *http.Client, index *Index) error {
	resp, err := client.Get(provider.ContentURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return ingestJSON(body, provider, index)
}

func ingestJSON(bytes []byte, provider *Provider, index *Index) error {
	var content []*Content
	err := json.Unmarshal(bytes, &content)
	if err != nil {
		return err
	}

	for _, item := range content {
		if len(item.Domains) == 0 {
			item.Domains = provider.Domains
		}
		item = maybeAppendExplanation(item)
	}

	index.Add(content)

	return nil
}

func ingestSyndicationFeed(provider *Provider, client *http.Client, index *Index) error {
	fp := gofeed.NewParser()
	fp.Client = client
	feed, err := fp.ParseURL(provider.ContentURL)
	if err != nil {
		return err
	}

	content := make([]*Content, 0)
	for _, item := range feed.Items {
		newc, err := createContentFromFeedItem(provider, item)
		if err != nil {
			return err
		}
		content = append(content, newc)
	}
	index.Add(content)

	return nil
}

func createContentFromFeedItem(provider *Provider, item *gofeed.Item) (*Content, error) {
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

	newc := &Content{
		ID:        findID(item),
		Source:    provider.ID,
		Title:     item.Title,
		URL:       item.Link,
		Image:     findImage(item, context),
		Excerpt:   summary,
		HTML:      item.Description,
		Tags:      append(item.Categories, provider.Categories...),
		Author:    processAuthor(item),
		Published: item.Published,
		Regions:   provider.Regions,
		Language:  provider.Language,
		Script:    provider.Script,
		Domains:   provider.Domains,
		CType:     RECOMMENDED}
	return maybeAppendExplanation(newc), nil
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

func maybeAppendExplanation(content *Content) *Content {
	if len(content.Tags) > 0 {
		content.Explanation = "Selected for users interested in " + strings.Join(content.Tags, ",")
	}
	return content
}
