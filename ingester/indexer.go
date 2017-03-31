package ingester

import (
	"log"
	"os"
	"path/filepath"

	"github.com/blevesearch/bleve"
	"github.com/nu7hatch/gouuid"
	"mozilla.org/crec/config"
	"mozilla.org/crec/content"
)

// Indexer responsible for indexing content
type Indexer struct {
	id         string
	content    []*content.Content
	contentMap map[string]*content.Content
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
