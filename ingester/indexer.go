package ingester

import "mozilla.org/crec/content"
import "github.com/blevesearch/bleve"
import "github.com/nu7hatch/gouuid"
import "path/filepath"
import "log"
import "os"

// Indexer responsible for indexing content
type Indexer struct {
	id         string
	content    []*content.Content
	contentMap map[string]*content.Content
	index      bleve.Index
}

const indexRoot = "index"
const indexPath = "crec.bleve"

// CreateIndexer create an instance of a content indexer
func CreateIndexer() *Indexer {
	u, err := uuid.NewV4()
	if err != nil {
		log.Fatal("Failed to create index directory:", err)
	}
	indexPath := filepath.FromSlash(indexRoot + "/" + u.String() + "/" + indexPath)
	index, err := bleve.Open(indexPath)
	if err != nil {
		mapping := bleve.NewIndexMapping()
		index, err = bleve.New(indexPath, mapping)
		if err != nil {
			log.Fatal("Failed to create index: ", err)
		}
	}

	return &Indexer{id: u.String(), content: make([]*content.Content, 0), index: index}
}

// RemoveAllIndexes deletes all existing indexes
func RemoveAllIndexes() error {
	err := os.RemoveAll(indexRoot)
	return err
}

// Add content to index
func (i *Indexer) Add(c *content.Content) error {
	if i.contentMap == nil {
		i.contentMap = make(map[string]*content.Content)
	}

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
