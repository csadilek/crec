package ingester

import "mozilla.org/crec/content"
import "github.com/blevesearch/bleve"
import "log"

// Indexer responsible for indexing content
type Indexer struct {
	Content    []*content.Content
	ContentMap map[string]*content.Content
	Index      bleve.Index
}

const path = "crec.bleve"

// CreateIndexer create an instance of a content indexer
func CreateIndexer() *Indexer {
	index, err := bleve.Open(path)
	if err != nil {
		mapping := bleve.NewIndexMapping()
		index, err = bleve.New(path, mapping)
		if err != nil {
			log.Fatal("Failed to create index: ", err)
		}
	}

	return &Indexer{Content: nil, Index: index}
}

// Add content to the index
func (i *Indexer) Add(c *content.Content) error {
	if i.ContentMap == nil {
		i.ContentMap = make(map[string]*content.Content)
	}
	var id string
	if c.Item.GUID != "" {
		id = c.Item.GUID
	} else {
		id = c.Item.Link
	}
	i.ContentMap[id] = c
	return i.Index.Index(id, c.Item.Description)
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
