package ingester

import (
	"log"
	"os"
	"path/filepath"

	"strings"

	"github.com/blevesearch/bleve"
	"github.com/nu7hatch/gouuid"
	"golang.org/x/text/language"
	"mozilla.org/crec/config"
	"mozilla.org/crec/content"
)

// Indexer responsible for indexing content
type Indexer struct {
	id         string
	content    []*content.Content
	contentMap map[string]*content.Content
	languages  map[string][]string
	regions    map[string][]string
	scripts    map[string][]string
	index      bleve.Index
}

// CreateIndexer create an instance of a content indexer
func CreateIndexer(indexRoot string, indexFile string) *Indexer {
	u, err := uuid.NewV4()
	if err != nil {
		log.Fatal("Failed to create index directory:", err)
	}
	indexPath := filepath.FromSlash(indexRoot + "/" + u.String() + "/" + indexFile)
	index, err := bleve.Open(indexPath)
	if err != nil {
		mapping := bleve.NewIndexMapping()
		index, err = bleve.New(indexPath, mapping)
		if err != nil {
			log.Fatal("Failed to create index: ", err)
		}
	}

	return &Indexer{
		id:         u.String(),
		content:    make([]*content.Content, 0),
		contentMap: make(map[string]*content.Content),
		languages:  make(map[string][]string),
		regions:    make(map[string][]string),
		scripts:    make(map[string][]string),
		index:      index}
}

// RemoveAll deletes all existing indexes
func RemoveAll(config *config.Config) error {
	err := os.RemoveAll(config.GetIndexDir())
	return err
}

// Add content to index
func (i *Indexer) Add(c *content.Content) error {
	i.content = append(i.content, c)
	i.contentMap[c.ID] = c

	if len(c.Regions) == 0 {
		i.regions["any"] = append(i.regions["any"], c.ID)
	} else {
		for _, region := range c.Regions {
			indexLocaleValue(region, c.ID, i.regions)
		}
	}
	indexLocaleValue(c.Language, c.ID, i.languages)
	indexLocaleValue(c.Script, c.ID, i.scripts)

	return i.index.Index(c.ID, c.Summary)
}

// Query index for content
func (i *Indexer) Query(q string) ([]*content.Content, error) {
	c := make([]*content.Content, 0)

	query := bleve.NewQueryStringQuery(q)
	searchRequest := bleve.NewSearchRequest(query)
	searchResult, err := i.index.Search(searchRequest)

	for _, hit := range searchResult.Hits {
		hitc := i.contentMap[hit.ID]
		if hitc != nil {
			c = append(c, hitc)
		}
	}
	return c, err
}

// GetID returns the unique ID of this indexer
func (i *Indexer) GetID() string {
	return i.id
}

// GetContent returns all indexed content
func (i *Indexer) GetContent() []*content.Content {
	return i.content
}

// GetLocalizedContent returns indexed content matching the provided language, script and regions
func (i *Indexer) GetLocalizedContent(tags []language.Tag) []*content.Content {
	if len(tags) == 0 {
		return i.content
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

	hitMap := make(map[string]int)
	for _, langHit := range langHits {
		hitMap[langHit]++
	}
	for _, regionHit := range regionHits {
		hitMap[regionHit]++
	}
	for _, scriptHit := range scriptHits {
		hitMap[scriptHit]++
		if hitMap[scriptHit] == 3 {
			c = append(c, i.contentMap[scriptHit])
		}
	}

	return c
}

func indexLocaleValue(key string, val string, m map[string][]string) {
	k := strings.ToLower(key)
	if k == "" {
		m["any"] = append(m["any"], val)
	} else {
		m[k] = append(m[k], val)
	}
}
