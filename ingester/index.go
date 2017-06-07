package ingester

import (
	"log"
	"path/filepath"
	"time"

	"strings"

	"github.com/blevesearch/bleve"
	"github.com/nu7hatch/gouuid"
	"golang.org/x/text/language"
	"mozilla.org/crec/config"
	"mozilla.org/crec/content"
)

// Index responsible for indexing content
type Index struct {
	id                   string
	allContent           []*content.Content
	content              map[string]*content.Content
	providers            map[string][]*content.Content
	providersLastUpdated map[string]time.Time
	languages            map[string][]*content.Content
	regions              map[string][]*content.Content
	scripts              map[string][]*content.Content
	fullText             bleve.Index
}

// CreateIndex creates an index instance, using the provided file name and root directory
func CreateIndex(c *config.Config) *Index {
	u, err := uuid.NewV4()
	if err != nil {
		log.Fatal("Failed to create index directory:", err)
	}

	var fullTextIndex bleve.Index
	if c.FullTextIndexActive() {
		indexPath := filepath.FromSlash(c.GetFullTextIndexDir() + "/" + u.String() + "/" + c.GetFullTextIndexFile())
		fullTextIndex, err = bleve.Open(indexPath)
		if err != nil {
			mapping := bleve.NewIndexMapping()
			fullTextIndex, err = bleve.New(indexPath, mapping)
			if err != nil {
				log.Fatal("Failed to create index: ", err)
			}
		}
	}

	return &Index{
		id:                   u.String(),
		allContent:           make([]*content.Content, 0),
		content:              make(map[string]*content.Content),
		providers:            make(map[string][]*content.Content),
		providersLastUpdated: make(map[string]time.Time),
		languages:            make(map[string][]*content.Content),
		regions:              make(map[string][]*content.Content),
		scripts:              make(map[string][]*content.Content),
		fullText:             fullTextIndex}
}

// AddAll adds the provided content items to this index
func (i *Index) AddAll(c []*content.Content) error {
	for _, content := range c {
		err := i.Add(content)
		if err != nil {
			return err
		}
	}
	return nil
}

// Add content to index
func (i *Index) Add(c *content.Content) error {
	i.allContent = append(i.allContent, c)
	i.content[c.ID] = c
	i.providers[c.Source] = append(i.providers[c.Source], c)

	if len(c.Regions) == 0 {
		i.regions["any"] = append(i.regions["any"], c)
	} else {
		for _, region := range c.Regions {
			indexLocaleValue(region, c, i.regions)
		}
	}
	indexLocaleValue(c.Language, c, i.languages)
	indexLocaleValue(c.Script, c, i.scripts)

	if i.fullText != nil {
		return i.fullText.Index(c.ID, c.Title+" "+c.Summary)
	}

	return nil
}

// Query index for content
func (i *Index) Query(q string) ([]*content.Content, error) {
	c := make([]*content.Content, 0)
	if i.fullText == nil {
		return c, nil
	}

	query := bleve.NewQueryStringQuery(q)
	searchRequest := bleve.NewSearchRequest(query)
	searchResult, err := i.fullText.Search(searchRequest)
	if searchResult != nil {
		for _, hit := range searchResult.Hits {
			hitc := i.content[hit.ID]
			if hitc != nil {
				c = append(c, hitc)
			}
		}
	}
	return c, err
}

// GetID returns the unique ID of this index
func (i *Index) GetID() string {
	return i.id
}

// GetContent returns all indexed content
func (i *Index) GetContent() []*content.Content {
	return i.allContent
}

// GetLocalizedContent returns indexed content matching the provided language, script and regions
func (i *Index) GetLocalizedContent(tags []language.Tag) []*content.Content {
	if len(tags) == 0 {
		return i.allContent
	}

	c := make([]*content.Content, 0)

	langHits := i.languages["any"]
	regionHits := i.regions["any"]
	scriptHits := i.scripts["any"]
	for _, tag := range tags {
		b, _ := tag.Base()
		r, _ := tag.Region()
		s, _ := tag.Script()
		langHits = append(langHits, i.languages[strings.ToLower(b.String())]...)
		regionHits = append(regionHits, i.regions[strings.ToLower(r.String())]...)
		scriptHits = append(scriptHits, i.scripts[strings.ToLower(s.String())]...)
	}

	hitMap := make(map[*content.Content]int)
	for _, langHit := range langHits {
		hitMap[langHit]++
	}
	for _, regionHit := range regionHits {
		hitMap[regionHit]++
	}
	for _, scriptHit := range scriptHits {
		hitMap[scriptHit]++
		if hitMap[scriptHit] == 3 {
			c = append(c, scriptHit)
		}
	}

	return c
}

// GetProviderLastUpdated returns the last updated time of the given provider
func (i *Index) GetProviderLastUpdated(provider string) time.Time {
	return i.providersLastUpdated[provider]
}

// SetProviderLastUpdated sets the last updated time of the given provider
func (i *Index) SetProviderLastUpdated(provider string) {
	i.providersLastUpdated[provider] = time.Now()
}

// GetProviderContent returns all indexed content from the given provider
func (i *Index) GetProviderContent(provider string) []*content.Content {
	return i.providers[provider]
}

func indexLocaleValue(key string, val *content.Content, m map[string][]*content.Content) {
	k := strings.ToLower(key)
	if k == "" {
		m["any"] = append(m["any"], val)
	} else {
		m[k] = append(m[k], val)
	}
}
