package ingester

import "mozilla.org/crec/content"
import "github.com/blevesearch/bleve"
import "github.com/nu7hatch/gouuid"
import "path/filepath"
import "log"

// Indexer responsible for indexing content
type Indexer struct {
	Content    []*content.Content
	ContentMap map[string]*content.Content
	Index      bleve.Index
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

	return &Indexer{Content: make([]*content.Content, 0), Index: index}
}

// Add content to the index
func (i *Indexer) Add(c *content.Content) error {
	if i.ContentMap == nil {
		i.ContentMap = make(map[string]*content.Content)
	}

	i.Content = append(i.Content, c)
	i.ContentMap[c.ID] = c
	return i.Index.Index(c.ID, c.Item.Description)
}

// Query the index for content
func (i *Indexer) Query(q string) ([]*content.Content, error) {
	c := make([]*content.Content, 0)

	query := bleve.NewQueryStringQuery(q)
	searchRequest := bleve.NewSearchRequest(query)
	searchResult, err := i.Index.Search(searchRequest)

	for _, hit := range searchResult.Hits {
		hitc := i.ContentMap[hit.ID]
		if hitc != nil {
			c = append(c, hitc)
		}
	}
	return c, err
}
