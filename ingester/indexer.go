package ingester

import (
	"log"
	"os"
	"path/filepath"

	"strings"

	"github.com/blevesearch/bleve"
	"github.com/nu7hatch/gouuid"
	"mozilla.org/crec/config"
	"mozilla.org/crec/content"
)

// Indexer responsible for indexing content
type Indexer struct {
	id          string
	content     []*content.Content
	contentMap  map[string]*content.Content
	languageMap map[string][]string
	regionMap   map[string][]string
	index       bleve.Index
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
		id:          u.String(),
		content:     make([]*content.Content, 0),
		contentMap:  make(map[string]*content.Content),
		languageMap: make(map[string][]string),
		regionMap:   make(map[string][]string),
		index:       index}
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

	// TODO default to all? q=?
	for _, region := range c.Regions {
		r := strings.ToLower(region)
		if _, ok := i.regionMap[r]; ok {
			i.regionMap[r] = append(i.regionMap[r], c.ID)
		} else {
			i.regionMap[r] = []string{c.ID}
		}
	}

	for _, language := range c.Languages {
		l := strings.ToLower(language)
		if _, ok := i.languageMap[l]; ok {
			i.languageMap[l] = append(i.languageMap[l], c.ID)
		} else {
			i.languageMap[l] = []string{c.ID}
		}
	}

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

// GetContent returns the indexed content
func (i *Indexer) GetContent() []*content.Content {
	return i.content
}
